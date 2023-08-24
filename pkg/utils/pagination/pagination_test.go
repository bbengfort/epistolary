package pagination_test

import (
	"testing"
	"time"

	. "github.com/bbengfort/epistolary/pkg/utils/pagination"
	"github.com/stretchr/testify/require"
)

func TestPaginationToken(t *testing.T) {
	cursor := New(1942, 2042, 100)
	token, err := cursor.PageToken()
	require.NoError(t, err, "could not create next page token")
	require.Greater(t, len(token), 32, "the token should be at least 32 characters")
	require.Less(t, len(token), 96, "the token should be no more than 96 characters")

	parsed, err := Parse(token)
	require.NoError(t, err, "could not parse token")
	require.Equal(t, cursor, parsed, "expected parsed token to be the same as the original")
}

func TestPaginationDefaultPageSize(t *testing.T) {
	cursor := New(0, 0, 0)
	require.Equal(t, DefaultPageSize, cursor.Size)
	require.False(t, cursor.IsZero(), "new cursor should not be zero valued")
}

func TestPaginationExpired(t *testing.T) {
	cursor := &Cursor{}
	expired, err := cursor.HasExpired()
	require.ErrorIs(t, err, ErrMissingExpiration)
	require.False(t, expired, "if err is not nil, expired should be false")

	cursor.Exp = time.Now().Add(5 * time.Minute).UnixMilli()
	expired, err = cursor.HasExpired()
	require.NoError(t, err, "cursor should compute expiration without error")
	require.False(t, expired, "cursor should not be expired for 5 minutes")

	cursor.Exp = time.Now().Add(-5 * time.Minute).UnixMilli()
	expired, err = cursor.HasExpired()
	require.NoError(t, err, "cursor should compute expiration without error")
	require.True(t, expired, "cursor should have expired 5 minutes ago")
}

func TestPaginationIsZero(t *testing.T) {
	cursor := &Cursor{}
	require.True(t, cursor.IsZero(), "empty cursor should be zero valued")

	cursor = New(1942, 2042, 100)
	require.False(t, cursor.IsZero(), "new cursor should not be zero valued")
}
