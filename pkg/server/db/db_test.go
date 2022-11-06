package db_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/bbengfort/epistolary/pkg/server/config"
	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/stretchr/testify/require"
)

// Connection to the test database used by all test functions.
var dsn = os.Getenv("EPISTOLARY_DATABASE_URL")

func TestConnectClose(t *testing.T) {
	if dsn == "" {
		t.Skip("no $EPISTOLARY_DATABASE_URL to connect to test database with")
	}

	conf := config.DatabaseConfig{
		URL:      dsn,
		ReadOnly: false,
	}

	// Try to open a transaction without connecting
	_, err := db.BeginTx(context.Background(), nil)
	require.EqualError(t, err, db.ErrNotConnected.Error())

	// Close the database without connecting
	err = db.Close()
	require.NoError(t, err, "close error when not connected")

	// Connect to the DB
	err = db.Connect(conf)
	require.NoError(t, err, "could not connect to db")

	// Connect to the DB again
	err = db.Connect(conf)
	require.NoError(t, err, "multiple connects causes error")

	// Open a transaction
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err, "could not create transaction")

	// Abort the transaction
	require.NoError(t, tx.Rollback(), "could not abort transaction")

	// Close connection to the DB
	require.NoError(t, db.Close(), "could not close db")

	// Reconnect to the DB
	require.NoError(t, db.Connect(conf), "could not reconnect to the db")
	require.NoError(t, db.Close(), "could not close db")
}

func TestReadOnly(t *testing.T) {
	if dsn == "" {
		t.Skip("no $EPISTOLARY_DATABASE_URL to connect to test database with")
	}

	conf := config.DatabaseConfig{
		URL:      dsn,
		ReadOnly: true,
	}

	// Ensure the DB is closed so it opens in readonly mode.
	require.NoError(t, db.Close(), "could not close db")

	// Connect to the DB in readonly mode
	require.NoError(t, db.Connect(conf), "could not connect to db")

	// Try a writable transaction
	_, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: false})
	require.EqualError(t, err, db.ErrReadOnly.Error())

	// Try a read only transaction
	_, err = db.BeginTx(context.Background(), nil)
	require.NoError(t, err, "couldn't create transaction from nil tx options")

	// Ensure the DB is closed when we're done
	require.NoError(t, db.Close(), "could not close db")
}
