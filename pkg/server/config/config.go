package config

import (
	"fmt"

	"github.com/bbengfort/epistolary/pkg/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

type Config struct {
	Maintenance  bool                `split_words:"true" default:"false"`
	BindAddr     string              `split_words:"true" default:":8000"`
	Mode         string              `split_words:"true" default:"release"`
	LogLevel     logger.LevelDecoder `split_words:"true" default:"info"`
	ConsoleLog   bool                `split_words:"true" default:"false"`
	AllowOrigins []string            `split_words:"true" default:"http://localhost:3000"`
	Database     DatabaseConfig
	processed    bool
}

type DatabaseConfig struct {
	URL      string `split_words:"true" required:"true"`
	ReadOnly bool   `split_words:"true" default:"false"`
	Testing  bool   `split_words:"true" default:"false"`
}

// New creates a new Config object from environment variables prefixed with EPISTOLARY.
func New() (conf Config, err error) {
	if err = envconfig.Process("epistolary", &conf); err != nil {
		return Config{}, err
	}

	// Validate the configuration
	if err = conf.Validate(); err != nil {
		return Config{}, err
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
	if err := c.Validate(); err != nil {
		return c, err
	}
	c.processed = true
	return c, nil
}

// Validate the config to make sure that it is usable to run the Epistolary server.
func (c Config) Validate() (err error) {
	if c.Mode != gin.ReleaseMode && c.Mode != gin.DebugMode && c.Mode != gin.TestMode {
		return fmt.Errorf("%q is not a valid gin mode", c.Mode)
	}
	return nil
}
