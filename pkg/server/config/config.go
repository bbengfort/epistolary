package config

import (
	"fmt"

	"github.com/bbengfort/epistolary/pkg"
	"github.com/bbengfort/epistolary/pkg/utils/logger"
	"github.com/bbengfort/epistolary/pkg/utils/sentry"
	"github.com/gin-gonic/gin"
	"github.com/rotationalio/confire"
	"github.com/rotationalio/confire/validate"
	"github.com/rs/zerolog"
)

type Config struct {
	Maintenance  bool                `split_words:"true" default:"false"`
	BindAddr     string              `split_words:"true" default:":8000"`
	Mode         GinMode             `split_words:"true" default:"release"`
	LogLevel     logger.LevelDecoder `split_words:"true" default:"info"`
	ConsoleLog   bool                `split_words:"true" default:"false"`
	AllowOrigins []string            `split_words:"true" default:"https://epistolary.app"`
	Database     DatabaseConfig
	Token        TokenConfig
	Sentry       sentry.Config
	processed    bool
}

type DatabaseConfig struct {
	URL      string `required:"true" split_words:"true"`
	ReadOnly bool   `split_words:"true" default:"false"`
	Testing  bool   `split_words:"true" default:"false"`
}

type TokenConfig struct {
	Keys         map[string]string `desc:"a map of key ID to key path"`
	Audience     string            `default:"https://epistolary.app"`
	Issuer       string            `default:"https://api.epistolary.app"`
	CookieDomain string            `split_words:"true" default:"epistolary.app"`
}

// New creates a new Config object from environment variables prefixed with EPISTOLARY.
func New() (conf Config, err error) {
	if err = confire.Process("epistolary", &conf); err != nil {
		return conf, err
	}

	// Ensure the Sentry release is named correctly
	if conf.Sentry.Release == "" {
		conf.Sentry.Release = fmt.Sprintf("epistolary@%s", pkg.Version())
	}

	conf.processed = true
	return conf, nil
}

func (c Config) GetLogLevel() zerolog.Level {
	return zerolog.Level(c.LogLevel)
}

func (c Config) IsZero() bool {
	return !c.processed
}

// Mark a manually constructed as processed as long as it is validated.
func (c Config) Mark() (Config, error) {
	if err := validate.Validate(&c); err != nil {
		return c, err
	}
	c.processed = true
	return c, nil
}

type GinMode string

func (s GinMode) Validate() error {
	if s != gin.ReleaseMode && s != gin.DebugMode && s != gin.TestMode {
		return fmt.Errorf("%q is not a valid gin mode", s)
	}
	return nil
}
