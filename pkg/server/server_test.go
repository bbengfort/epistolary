package server_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server"
	"github.com/bbengfort/epistolary/pkg/server/config"
	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/bbengfort/epistolary/pkg/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type epistolaryTestSuite struct {
	suite.Suite
	srv    *server.Server
	conf   config.Config
	client api.EpistolaryClient
	dbPath string
	stop   chan bool
}

// Run once before all the tests are executed
func (suite *epistolaryTestSuite) SetupSuite() {
	require := suite.Require()
	suite.stop = make(chan bool, 1)

	// Create a temporary test database for the tests
	var err error
	suite.dbPath, err = os.MkdirTemp("", "epistolary-*")
	require.NoError(err, "could not create temporary directory for database")

	// Create a test configuration to run the Epistolary API server as a fully
	// functional server on an open port using the local-loopback for networking.
	suite.conf, err = config.Config{
		Maintenance:  false,
		BindAddr:     "127.0.0.1:0",
		Mode:         gin.TestMode,
		LogLevel:     logger.LevelDecoder(zerolog.DebugLevel),
		ConsoleLog:   false,
		AllowOrigins: []string{"http://localhost:3000"},
		Database: config.DatabaseConfig{
			ReadOnly: false,
			Testing:  true,
		},
		Token: config.TokenConfig{
			Keys: map[string]string{
				"01GE6191AQTGMCJ9BN0QC3CCVG": "testdata/01GE6191AQTGMCJ9BN0QC3CCVG.pem",
				"01GE62EXXR0X0561XD53RDFBQJ": "testdata/01GE62EXXR0X0561XD53RDFBQJ.pem",
			},
			Audience: "http://localhost:3000",
			Issuer:   "http://localhost:8000",
		},
	}.Mark()
	require.NoError(err, "test configuration is invalid")

	suite.srv, err = server.New(suite.conf)
	require.NoError(err, "could not create the epistolary api server from the test configuration")

	// Start the BFF server - the goal of the tests is to have the server run for the
	// entire duration of the tests. Implement reset methods to ensure the server state
	// doesn't change between tests in Before/After.
	go func() {
		suite.srv.Serve()
		suite.stop <- true
	}()

	// Wait for 500ms to ensure the API server starts up
	time.Sleep(500 * time.Millisecond)

	// Create a Epistolary client for making requests to the server
	require.NotEmpty(suite.srv.URL(), "no url to connect the client on")
	suite.client, err = api.New(suite.srv.URL())
	require.NoError(err, "could not initialize the epistolary client")
}

func (suite *epistolaryTestSuite) TearDownSuite() {
	require := suite.Require()

	// Set the db mock to expect close
	db.Mock().ExpectClose()

	// Shutdown the API server
	err := suite.srv.Shutdown()
	require.NoError(err, "could not gracefully shutdown the epistolary test server")

	// Wait for server to stop to prevent race conditions
	<-suite.stop

	// Cleanup temporary test directory
	err = os.RemoveAll(suite.dbPath)
	require.NoError(err, "could not cleanup temporary database")
}

func (suite *epistolaryTestSuite) ResetDatabase() (err error) {
	// Truncate all database tables except roles, permissions, and role_permissions
	stmts := []string{
		"TRUNCATE users",
		"TRUNCATE user_roles",
		"DELETE FROM api_key_permissions",
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(context.Background(), nil); err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range stmts {
		if _, err = tx.Exec(stmt); err != nil {
			return err
		}
	}

	tx.Commit()
	return nil
}

func TestEpistolary(t *testing.T) {
	suite.Run(t, &epistolaryTestSuite{})
}
