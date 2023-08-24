package pagination_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	. "github.com/bbengfort/epistolary/pkg/utils/pagination"
	"github.com/stretchr/testify/require"
)

func TestCursor(t *testing.T) {
	type testCase struct {
		start   int64
		end     int64
		size    uint32
		expires int64
	}

	testCases := []testCase{
		{0, 0, 0, 0},
		{0, 100, 100, time.Now().UnixMilli()},
		{math.MaxInt64, math.MaxInt64, math.MaxUint32, math.MaxInt64},
	}

	// Append random test cases
	for i := 0; i < 100; i++ {
		tc := testCase{
			start:   rand.Int63(),
			end:     rand.Int63(),
			size:    rand.Uint32(),
			expires: rand.Int63(),
		}
		testCases = append(testCases, tc)
	}

	for i, tc := range testCases {
		original := Cursor{tc.start, tc.end, tc.size, tc.expires}

		data, err := original.MarshalText()
		require.NoError(t, err, "could not encode test case %d cursor", i)

		require.Greater(t, len(data), 32, "the token should be at least 32 characters")
		require.Less(t, len(data), 128, "the token should be no more than 128 characters")

		actual := Cursor{}
		err = actual.UnmarshalText(data)
		require.NoError(t, err, "could not decode test case %d", i)

		require.Equal(t, original, actual, "cursors do not match")
	}
}

func TestCursorExpires(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	cursor := Cursor{0, 100, 100, now.UnixMilli()}
	require.Equal(t, now, cursor.Expires())
}
