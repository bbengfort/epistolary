package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bbengfort/epistolary/pkg/server/config"
	_ "github.com/lib/pq"
)

var (
	readonly bool            // if true, only allow database reads
	conn     *sql.DB         // connection pool to the DB managed by the package
	connmu   sync.RWMutex    // synchronize connect and close DB connection
	connect  sync.Once       // ensure that the database is only connected to once
	mock     sqlmock.Sqlmock // mock for unit testing modules that use the database
)

var (
	ErrNotConnected = errors.New("not connected to the database")
	ErrReadOnly     = errors.New("connected in readonly mode - only readonly transactions allowed")
	ErrNotfound     = errors.New("record not found or no rows returned")
)

// Connect to the Postgres stabase specified by theDSN. Connecting in read-only mode is
// is managed by the package, not by the database. Multiple or concurrent calls to
// Connect will be ignored even if a different configuration is passed.
func Connect(conf config.DatabaseConfig) (err error) {
	// Guard against concurrent Connect and Close
	connmu.Lock()

	// Ensure that the connect function is only called once.
	connect.Do(func() {
		if conf.Testing {
			// Create a sqlmock for testing purposes
			readonly = conf.ReadOnly
			conn, mock, err = sqlmock.New()
			return
		}

		readonly = conf.ReadOnly
		if conn, err = sql.Open("postgres", conf.URL); err != nil {
			return
		}

		conn.SetMaxOpenConns(16)
		conn.SetMaxIdleConns(8)
		conn.SetConnMaxLifetime(90 * time.Minute)
		conn.SetConnMaxIdleTime(90 * time.Second)

		if err = conn.Ping(); err != nil {
			return
		}
	})

	connmu.Unlock()
	return err
}

// Close the database safely and allow reconnect after close.
func Close() (err error) {
	connmu.Lock()
	if conn != nil {
		err = conn.Close()
		conn = nil
		mock = nil
		connect = sync.Once{}
	}
	connmu.Unlock()
	return err
}

// BeginTx creates a transaction with the connected dtabase; errors if not connected.
func BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error) {
	connmu.RLock()
	if conn == nil {
		connmu.RUnlock()
		return nil, ErrNotConnected
	}

	if opts == nil {
		opts = &sql.TxOptions{ReadOnly: readonly}
	} else if readonly && !opts.ReadOnly {
		connmu.RUnlock()
		return nil, ErrReadOnly
	}

	tx, err = conn.BeginTx(ctx, opts)
	connmu.RUnlock()
	return tx, err
}

// Mock returns the sql mock object for use in unit testing
func Mock() sqlmock.Sqlmock {
	return mock
}
