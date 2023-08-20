package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bbengfort/epistolary/pkg/server/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var testEnv = map[string]string{
	"EPISTOLARY_MAINTENANCE":              "false",
	"EPISTOLARY_BIND_ADDR":                "8080",
	"EPISTOLARY_MODE":                     "debug",
	"EPISTOLARY_LOG_LEVEL":                "debug",
	"EPISTOLARY_CONSOLE_LOG":              "true",
	"EPISTOLARY_ALLOW_ORIGINS":            "https://epistolary.app",
	"EPISTOLARY_DATABASE_URL":             "postgres://localhost:5432/epistolary?sslmode=disable",
	"EPISTOLARY_DATABASE_READ_ONLY":       "true",
	"EPISTOLARY_DATABASE_TESTING":         "true",
	"EPISTOLARY_TOKEN_KEYS":               "01GECSDK5WJ7XWASQ0PMH6K41K:testdata/01GECSDK5WJ7XWASQ0PMH6K41K.pem,01GECSJGDCDN368D0EENX23C7R:testdata/01GECSJGDCDN368D0EENX23C7R.pem",
	"EPISTOLARY_TOKEN_AUDIENCE":           "http://localhost:3000",
	"EPISTOLARY_TOKEN_ISSUER":             "http://localhost:8000",
	"EPISTOLARY_TOKEN_COOKIE_DOMAIN":      "localhost",
	"EPISTOLARY_SENTRY_DSN":               "http://testing.sentry.test/1234",
	"EPISTOLARY_SENTRY_SERVER_NAME":       "tnode",
	"EPISTOLARY_SENTRY_ENVIRONMENT":       "testing",
	"EPISTOLARY_SENTRY_RELEASE":           "", // This should always be empty!
	"EPISTOLARY_SENTRY_TRACK_PERFORMANCE": "true",
	"EPISTOLARY_SENTRY_SAMPLE_RATE":       "0.95",
	"EPISTOLARY_SENTRY_DEBUG":             "true",
}

func TestConfig(t *testing.T) {
	// Set required environment variables and cleanup after
	prevEnv := curEnv()
	t.Cleanup(func() {
		for key, val := range prevEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	})
	setEnv()

	conf, err := config.New()
	require.NoError(t, err)
	require.False(t, conf.IsZero())

	require.False(t, conf.Maintenance)
	require.Equal(t, testEnv["EPISTOLARY_BIND_ADDR"], conf.BindAddr)
	require.Equal(t, testEnv["EPISTOLARY_MODE"], conf.Mode)
	require.Equal(t, zerolog.DebugLevel, conf.GetLogLevel())
	require.True(t, conf.ConsoleLog)
	require.Equal(t, testEnv["EPISTOLARY_DATABASE_URL"], conf.Database.URL)
	require.True(t, conf.Database.ReadOnly)
	require.True(t, conf.Database.Testing)
	require.Len(t, conf.AllowOrigins, 1)
	require.Len(t, conf.Token.Keys, 2)
	require.Equal(t, testEnv["EPISTOLARY_TOKEN_AUDIENCE"], conf.Token.Audience)
	require.Equal(t, testEnv["EPISTOLARY_TOKEN_ISSUER"], conf.Token.Issuer)
	require.Equal(t, testEnv["EPISTOLARY_TOKEN_COOKIE_DOMAIN"], conf.Token.CookieDomain)
	require.Equal(t, testEnv["EPISTOLARY_SENTRY_DSN"], conf.Sentry.DSN)
	require.Equal(t, testEnv["EPISTOLARY_SENTRY_SERVER_NAME"], conf.Sentry.ServerName)
	require.Equal(t, testEnv["EPISTOLARY_SENTRY_ENVIRONMENT"], conf.Sentry.Environment)
	require.True(t, conf.Sentry.TrackPerformance)
	require.Equal(t, 0.95, conf.Sentry.SampleRate)
	require.True(t, conf.Sentry.Debug)

	// Ensure the sentry release is configured correctly
	require.True(t, strings.HasPrefix(conf.Sentry.GetRelease(), "epistolary@"))
}

func TestRequiredConfig(t *testing.T) {
	required := []string{
		"EPISTOLARY_DATABASE_URL",
	}

	// Collect required environment variables and cleanup after
	prevEnv := curEnv(required...)
	cleanup := func() {
		for key, val := range prevEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
	t.Cleanup(cleanup)

	// Ensure that we've captured the complete set of required environment variables
	setEnv(required...)
	_, err := config.New()
	require.NoError(t, err)

	// Ensure that each environment variable is required
	for _, envvar := range required {
		// Add all environment variables but the current one
		for _, key := range required {
			if key == envvar {
				os.Unsetenv(key)
			} else {
				setEnv(key)
			}
		}

		_, err := config.New()
		require.Errorf(t, err, "expected %q to be required but no error occurred", envvar)
	}
}

// Returns the current environment for the specified keys, or if no keys are specified
// then returns the current environment for all keys in testEnv.
func curEnv(keys ...string) map[string]string {
	env := make(map[string]string)
	if len(keys) > 0 {
		for _, envvar := range keys {
			if val, ok := os.LookupEnv(envvar); ok {
				env[envvar] = val
			}
		}
	} else {
		for key := range testEnv {
			env[key] = os.Getenv(key)
		}
	}

	return env
}

// Sets the environment variable from the testEnv, if no keys are specified, then sets
// all environment variables from the test env.
func setEnv(keys ...string) {
	if len(keys) > 0 {
		for _, key := range keys {
			if val, ok := testEnv[key]; ok {
				os.Setenv(key, val)
			}
		}
	} else {
		for key, val := range testEnv {
			os.Setenv(key, val)
		}
	}
}
