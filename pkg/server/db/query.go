package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// Prep converts named query arguments into positional arguments for postgres queries.
func Prep(query string, args ...any) (string, []any) {
	var keys []string
	var vals []any

	for i, a := range args {
		if na, ok := a.(sql.NamedArg); ok {
			keys = append(keys, fmt.Sprintf(":%s", na.Name))
			keys = append(keys, fmt.Sprintf("$%d", i+1))
			vals = append(vals, na.Value)
		} else {
			vals = append(vals, a)
		}
	}

	query = strings.NewReplacer(keys...).Replace(query)
	return query, vals
}
