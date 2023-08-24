package db_test

import (
	"database/sql"
	"testing"

	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/stretchr/testify/require"
)

func TestPrep(t *testing.T) {
	testCases := []struct {
		query    string
		args     []any
		expected string
		params   []any
	}{
		{
			"SELECT * FROM readings ORDER BY created DESC LIMIT 10",
			nil,
			"SELECT * FROM readings ORDER BY created DESC LIMIT 10",
			nil,
		},
		{
			"SELECT * FROM readings WHERE user_id=:userID AND created > :after LIMIT 10",
			[]any{sql.Named("userID", 1), sql.Named("after", "2022-03-01T12:00:00Z")},
			"SELECT * FROM readings WHERE user_id=$1 AND created > $2 LIMIT 10",
			[]any{1, "2022-03-01T12:00:00Z"},
		},
		{
			"SELECT * FROM readings WHERE user_id=$1 AND created > $2 LIMIT 10",
			[]any{1, "2022-03-01T12:00:00Z"},
			"SELECT * FROM readings WHERE user_id=$1 AND created > $2 LIMIT 10",
			[]any{1, "2022-03-01T12:00:00Z"},
		},
		{
			"SELECT * FROM readings WHERE (user_id=:userID AND organization_id=:orgID) OR (user_id=:userID AND no_organization is :noOrg)",
			[]any{sql.Named("noOrg", true), sql.Named("userID", 1), sql.Named("orgID", 4)},
			"SELECT * FROM readings WHERE (user_id=$2 AND organization_id=$3) OR (user_id=$2 AND no_organization is $1)",
			[]any{true, 1, 4},
		},
	}

	for i, tc := range testCases {
		actual, params := db.Prep(tc.query, tc.args...)
		require.Equal(t, tc.expected, actual, "test case %d failed", i)
		require.Equal(t, tc.params, params, "test case %d failed", i)
	}
}
