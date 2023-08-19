package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GinLogger returns a new Gin middleware that performs logging for our JSON APIs using
// zerolog rather than the default Gin logger which is a standard HTTP logger.
// NOTE: we previously used github.com/dn365/gin-zerolog but wanted more customization.
func GinLogger(server string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Before request
		started := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// Handle the request
		c.Next()

		// After request
		status := c.Writer.Status()
		logctx := log.With().
			Strs("errors", c.Errors.Errors()).
			Str("path", path).
			Str("ser_name", server).
			Str("method", c.Request.Method).
			Dur("resp_time", time.Since(started)).
			Int("resp_bytes", c.Writer.Size()).
			Int("status", status).
			Str("client_ip", c.ClientIP()).
			Str("request_id", c.GetString("request_id")).
			Logger()

		msg := fmt.Sprintf("%s %s %s %d", server, c.Request.Method, c.Request.URL.Path, status)

		switch {
		case status >= 400 && status < 500:
			logctx.Warn().Msg(msg)
		case status >= 500:
			logctx.Error().Msg(msg)
		default:
			logctx.Info().Msg(msg)
		}
	}
}
