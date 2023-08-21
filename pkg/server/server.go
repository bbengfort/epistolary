package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/bbengfort/epistolary/pkg"
	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server/config"
	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/bbengfort/epistolary/pkg/server/db/schema"
	"github.com/bbengfort/epistolary/pkg/server/tokens"
	"github.com/bbengfort/epistolary/pkg/utils/logger"
	"github.com/bbengfort/epistolary/pkg/utils/sentry"
)

func init() {
	// Initialize zerolog with GCP logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = logger.GCPFieldKeyTime
	zerolog.MessageFieldName = logger.GCPFieldKeyMsg

	// Add the severity hook for GCP logging
	var gcpHook logger.SeverityHook
	log.Logger = zerolog.New(os.Stdout).Hook(gcpHook).With().Timestamp().Logger()
}

type Server struct {
	sync.RWMutex
	conf    config.Config
	srv     *http.Server
	router  *gin.Engine
	tokens  *tokens.TokenManager
	started time.Time
	healthy bool
	url     string
	errc    chan error
}

// New creates a new Epistolary server from the specified configuration.
func New(conf config.Config) (s *Server, err error) {
	// Load the default configuration from the environment if config is empty
	if conf.IsZero() {
		if conf, err = config.New(); err != nil {
			return nil, err
		}
	}

	// Set the global level
	zerolog.SetGlobalLevel(conf.GetLogLevel())

	// Set human readable logging if specified
	if conf.ConsoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Configure sentry for error and performance monitoring
	if conf.Sentry.UseSentry() {
		if err = sentry.Init(conf.Sentry); err != nil {
			return nil, err
		}
	}

	// Create the server and prepare to serve
	s = &Server{
		conf: conf,
		errc: make(chan error, 1),
	}

	// Connect to the TestNet and MainNet directory services and database if we're not
	// in maintenance or testing mode (in testing mode, the connection will be manual).
	if !s.conf.Maintenance {
		log.Debug().Msg("setting up production mode")
		if len(s.conf.Token.Keys) == 0 {
			return nil, errors.New("invalid configuration: no token keys specified")
		}

		if s.tokens, err = tokens.New(s.conf.Token); err != nil {
			return nil, err
		}
	}

	// Create the router
	gin.SetMode(string(conf.Mode))
	s.router = gin.New()
	if err = s.setupRoutes(); err != nil {
		return nil, err
	}

	// Create the http server
	s.srv = &http.Server{
		Addr:         s.conf.BindAddr,
		Handler:      s.router,
		ErrorLog:     nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s, nil
}

// Serve API requests on the specified address.
func (s *Server) Serve() (err error) {
	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.errc <- s.Shutdown()
	}()

	if s.conf.Maintenance {
		log.Warn().Msg("starting server in maintenance mode")
	} else {
		// Wait for the database to be at the correct schema
		// TODO: handle testing mode a bit more gracefully
		if !s.conf.Database.Testing {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			err = schema.Wait(ctx, s.conf.Database.URL)
			cancel()
			if err != nil {
				return err
			}
		}

		// Connect to the database
		if err = db.Connect(s.conf.Database); err != nil {
			return err
		}
		log.Debug().Bool("read-only", s.conf.Database.ReadOnly).Str("dsn", s.conf.Database.URL).Msg("connected to database")
	}

	// Set the health of the service to true unless we're in maintenance mode.
	// The server should still start so that it can return unavailable to requests.
	s.SetHealth(!s.conf.Maintenance)

	// Create a socket to listen on so that we can infer the final URL (e.g. if the
	// BindAddr is 127.0.0.1:0 for testing, a random port will be assigned, manually
	// creating the listener will allow us to determine which port).
	var sock net.Listener
	if sock, err = net.Listen("tcp", s.conf.BindAddr); err != nil {
		s.errc <- err
	}

	// Set the URL from the listener
	s.SetURL("http://" + sock.Addr().String())
	s.started = time.Now()

	// Listen for HTTP requests on the specified address and port
	go func() {
		if err = s.srv.Serve(sock); err != nil && err != http.ErrServerClosed {
			s.errc <- err
		}

		// If there is no error, return nil so this function exits if Shutdown is
		// called manually (e.g. not from an OS signal).
		s.errc <- nil
	}()

	log.Info().
		Str("listen", s.url).
		Str("version", pkg.Version()).
		Msg("epistolary server started")

	// Listen for any errors that might have occurred and wait for all go routines to stop
	if err = <-s.errc; err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown() (err error) {
	log.Info().Msg("gracefully shutting down")

	s.SetHealth(false)
	s.srv.SetKeepAlivesEnabled(false)

	// Require shutdown in 30 seconds without blocking
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if serr := s.srv.Shutdown(ctx); serr != nil {
		err = multierror.Append(err, serr)
	}

	if !s.conf.Maintenance {
		if serr := db.Close(); serr != nil {
			err = multierror.Append(err, serr)
		}
	}

	// Flush sentry errors
	if s.conf.Sentry.UseSentry() {
		sentry.Flush(2 * time.Second)
	}

	if err == nil {
		log.Debug().Msg("successfully shutdown server")
	}
	return err
}

func (s *Server) SetHealth(health bool) {
	s.Lock()
	s.healthy = health
	s.Unlock()
	log.Debug().Bool("healthy", health).Msg("server health set")
}

func (s *Server) SetURL(url string) {
	s.Lock()
	s.url = url
	s.Unlock()
	log.Debug().Str("url", url).Msg("server url set")
}

func (s *Server) URL() string {
	s.RLock()
	defer s.RUnlock()
	return s.url
}

func (s *Server) setupRoutes() (err error) {
	// Instantiate Sentry Handlers
	var tags gin.HandlerFunc
	if s.conf.Sentry.UseSentry() {
		tagmap := map[string]string{"service": "epistolary", "component": "api"}
		tags = sentry.UseTags(tagmap)
	}

	var tracing gin.HandlerFunc
	if s.conf.Sentry.UsePerformanceTracking() {
		tagmap := map[string]string{"service": "epistolary", "component": "api"}
		tracing = sentry.TrackPerformance(tagmap)
	}

	// Application Middleware
	// NOTE: ordering is very important to how middleware is handled.
	middlewares := []gin.HandlerFunc{
		// Logging should be outside so we can record the complete latency of requests.
		// NOTE: logging panics will not recover.
		logger.GinLogger("epistolary"),

		// Panic recovery middleware
		// NOTE: gin middleware needs to be added before sentry
		gin.Recovery(),
		sentrygin.New(sentrygin.Options{
			Repanic:         true,
			WaitForDelivery: false,
		}),

		// Add searchable tags to sentry context
		tags,

		// Tracing helps us measure performance metrics with Sentry
		tracing,

		// CORS configuration allows the front-end to make cross-origin requests.
		cors.New(cors.Config{
			AllowOrigins:     s.conf.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-CSRF-TOKEN", "sentry-trace", "baggage"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),

		// Maintenance mode handling - does not require authentication.
		s.Available(),
	}

	// Add the middleware to the router
	for _, middleware := range middlewares {
		if middleware != nil {
			s.router.Use(middleware)
		}
	}

	// Add the v1 API routes
	v1 := s.router.Group("/v1")
	{
		// Registration route (no authentication required)
		v1.POST("/register", s.Register)

		// Login route (no authentication required)
		v1.POST("/login", s.Login)
		v1.POST("/logout", s.Logout)

		// Reading REST Resource (requires authentication)
		r := v1.Group("/reading", s.Authenticate)
		{
			r.GET("", s.Authorize("epistles:read"), s.ListReadings)
			r.POST("", s.Authorize("epistles:update"), s.CreateReading)
			r.GET("/:readingID", s.Authorize("epistles:read"), s.FetchReading)
			r.PUT("/:readingID", s.Authorize("epistles:update"), s.UpdateReading)
			r.DELETE("/:readingID", s.Authorize("epistles:delete"), s.DeleteReading)
		}

		// Heartbeat route (no authentication required)
		v1.GET("/status", s.Status)
	}

	// The "well known" routes expose client security information and credentials.
	wk := s.router.Group("/.well-known")
	{
		wk.GET("/jwks.json", s.JWKS)
		wk.GET("/security.txt", s.SecurityTxt)
		wk.GET("/openid-configuration", s.OpenIDConfiguration)
	}

	// NotFound and NotAllowed routes
	s.router.NoRoute(api.NotFound)
	s.router.NoMethod(api.NotAllowed)
	return nil
}
