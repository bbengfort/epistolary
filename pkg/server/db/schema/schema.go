package schema

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bbengfort/epistolary/pkg/utils/sentry"
	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/rs/zerolog/log"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

var (
	migrator *migrate.Migrate
	driver   source.Driver
	ready    bool
	readymu  sync.Mutex
)

var (
	ErrDirtySchema    = errors.New("schema is dirty requiring a manual migration")
	ErrVariadicDSN    = errors.New("specify zero or 1 dsn urls for configuration")
	ErrNoDatabaseURL  = errors.New("improperly configured: no database dsn")
	ErrNotInitialized = errors.New("schema package not initialized")
)

type Version struct {
	Required uint `json:"required,omitempty"`
	Current  uint `json:"current,omitempty"`
	Dirty    bool `json:"dirty,omitempty"`
}

// Configure the schema package to perform migration work. This must be called first.
func Configure(vdsn ...string) (err error) {
	readymu.Lock()
	defer readymu.Unlock()

	// If ready, don't reconfigure the migrator
	if ready {
		return nil
	}

	var dsn string
	switch len(vdsn) {
	case 0:
		if dsn, err = findDatabaseDSN(); err != nil {
			return err
		}
	case 1:
		dsn = vdsn[0]
	default:
		return ErrVariadicDSN
	}

	if dsn == "" {
		return ErrNoDatabaseURL
	}

	source := bindata.Resource(AssetNames(), func(name string) ([]byte, error) {
		return Asset(name)
	})

	if driver, err = bindata.WithInstance(source); err != nil {
		return err
	}

	if migrator, err = migrate.NewWithSourceInstance("go-bindata", driver, dsn); err != nil {
		return err
	}

	ready = true
	return nil
}

// Close the migrator and any open sources (errors are ignored)
func Close() {
	readymu.Lock()
	if migrator != nil {
		migrator.Close()
	}

	migrator = nil
	driver = nil
	ready = false
	readymu.Unlock()
}

// Current version returns the current state of the database
func CurrentVersion(dsn ...string) (*Version, error) {
	if err := Configure(dsn...); err != nil {
		return nil, err
	}

	current, dirty, _ := migrator.Version()
	return &Version{
		Required: findRequiredVersion(),
		Current:  current,
		Dirty:    dirty,
	}, nil
}

// Verify checks to ensure the database is ready with schema embedded in the binary
func Verify(dsn ...string) (err error) {
	var vers *Version
	if vers, err = CurrentVersion(dsn...); err != nil {
		return err
	}

	if vers.Required != vers.Current {
		return fmt.Errorf("schema version mismatch: current %d required %d", vers.Current, vers.Required)
	}

	if vers.Dirty {
		return ErrDirtySchema
	}
	return nil
}

// Migrate the database to the latest schema version
func Migrate(dsn ...string) (err error) {
	if err = Configure(dsn...); err != nil {
		return err
	}

	if err = migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil
}

// Down moves the database active schema version all the way down to 0
func Down(dsn ...string) (err error) {
	if err = Configure(dsn...); err != nil {
		return err
	}

	if err = migrator.Down(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil
}

// Drop deletes everything in the database
func Drop(dsn ...string) (err error) {
	if err = Configure(dsn...); err != nil {
		return err
	}

	if err = migrator.Drop(); err != nil {
		return err
	}
	return nil
}

// Force sets the database to the current required version regardless of the current
// active state and resets the dirty flag.
func Force(dsn ...string) (err error) {
	if err = Configure(dsn...); err != nil {
		return err
	}

	vers := findRequiredVersion()
	if err = migrator.Force(int(vers)); err != nil {
		return err
	}
	return nil
}

// Wait until the schema has been migrated to the required version or until the
// context is canceled, whichever comes first.
func Wait(ctx context.Context, dsn ...string) (err error) {
	var iters uint32
	for {
		if err = Verify(dsn...); err == nil {
			vers, _ := CurrentVersion(dsn...)
			log.Debug().Uint("version", vers.Current).Bool("dirty", vers.Dirty).Msg("at required schema version")
			return nil
		}

		// Log every minute rather than every 5 seconds
		if iters%12 == 0 {
			sentry.Warn(ctx).Err(err).Msg("waiting for correct schema version")
		} else {
			log.Trace().Err(err).Msg("waiting for correct schema version")
		}

		// Wait 5 seconds or until context deadline exceeded
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return fmt.Errorf("wait for correct schema canceled: %s", err)
		case <-time.After(5 * time.Second):
		}

		iters++
	}
}

func findDatabaseDSN() (string, error) {
	// Search for the DSN in the environment
	for _, envvar := range []string{"COSMOS_DATABASE_URL", "DATABASE_URL"} {
		if dsn := os.Getenv(envvar); dsn != "" {
			return dsn, nil
		}
	}
	return "", ErrNoDatabaseURL
}

func findRequiredVersion() uint {
	var (
		version uint
		next    uint
		err     error
	)

	if version, err = driver.First(); err != nil {
		return version
	}

	for {
		if next, err = driver.Next(version); err != nil {
			break
		}
		version = next
	}
	return version
}
