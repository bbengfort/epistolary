package epistles

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/bbengfort/epistolary/pkg/server/users"
	"github.com/bbengfort/epistolary/pkg/utils/pagination"
)

// Status constants
type Status string

const (
	StatusQueued   Status = "queued"
	StatusStarted  Status = "started"
	StatusFinished Status = "finished"
	StatusArchived Status = "archived"
)

// Database model for a reading object.
type Reading struct {
	EpistleID int64
	UserID    int64
	Status    Status
	Started   sql.NullTime
	Finished  sql.NullTime
	Archived  sql.NullTime
	Created   time.Time
	Modified  time.Time
	epistle   *Epistle
	user      *users.User
}

const (
	createReadingSQL = "INSERT INTO reading (epistle_id, user_id) VALUES ($1, $2)"
	readingTSSQL     = "SELECT created, modified FROM reading WHERE epistle_id=$1 AND user_id=$2"
)

// Create a reading for a user with a link.
func Create(ctx context.Context, userID int64, link string) (r *Reading, err error) {
	r = &Reading{UserID: userID}
	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get or create the epistle for the reading
	if r.epistle, err = getOrCreateEpistle(tx, link); err != nil {
		return nil, err
	}
	r.EpistleID = r.epistle.ID

	// Insert the reading into the database
	if _, err = tx.Exec(createReadingSQL, r.EpistleID, r.UserID); err != nil {
		return nil, err
	}

	// Get the timestamps from the database
	if err = tx.QueryRow(readingTSSQL, r.EpistleID, r.UserID).Scan(&r.Created, &r.Modified); err != nil {
		return nil, err
	}

	tx.Commit()
	return r, nil
}

const (
	countReadingSQL = "SELECT count(epistle_id) FROM reading WHERE user_id=$1"
	listReadingSQL  = "SELECT r.epistle_id, r.status, e.id, e.link, e.title, e.favicon, r.created, r.modified FROM reading r JOIN epistles e ON r.epistle_id=e.id"
)

// List readings for the specified user.
func List(ctx context.Context, userID int64, prevPage *pagination.Cursor) (r []*Reading, cursor *pagination.Cursor, err error) {
	if prevPage == nil {
		prevPage = pagination.New(0, 0, 0)
	}

	if prevPage.Size <= 0 {
		return nil, nil, ErrMissingPageSize
	}

	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	// Build paramaterized query with WHERE clause
	var query strings.Builder
	query.WriteString(listReadingSQL)

	params := make([]any, 0, 3)
	where := make([]string, 0, 2)

	params = append(params, sql.Named("userID", userID))
	where = append(where, "r.user_id=:userID")

	if prevPage.End != 0 {
		params = append(params, sql.Named("endIndex", prevPage.End))
		where = append(where, "r.epistle_id < :endIndex")
	}

	// Add the where clause to the query
	query.WriteString(" WHERE ")
	query.WriteString(strings.Join(where, " AND "))

	// Sort results by descending created timestamp
	query.WriteString(" ORDER BY created DESC")

	// Add the limit as the page size + 1 to perform a has next page check
	params = append(params, sql.Named("pageSize", prevPage.Size+1))
	query.WriteString(" LIMIT :pageSize")

	// Prep the query to convert named arguments into positional arguments
	qs, args := db.Prep(query.String(), params...)

	var rows *sql.Rows
	if rows, err = tx.Query(qs, args...); err != nil {
		return nil, nil, err
	}

	nRows := uint32(0)
	r = make([]*Reading, 0, prevPage.Size)
	defer rows.Close()
	for rows.Next() {
		// The query will request one additional message past the page size to check if
		// there is a next page. No rows should be processed after the page size.
		nRows++
		if nRows > prevPage.Size {
			continue
		}

		reading := &Reading{
			UserID: userID,
		}
		epistle := &Epistle{}

		if err = rows.Scan(
			&reading.EpistleID,
			&reading.Status,
			&epistle.ID,
			&epistle.Link,
			&epistle.Title,
			&epistle.Favicon,
			&reading.Created,
			&reading.Modified); err != nil {
			return nil, nil, err
		}

		reading.epistle = epistle
		r = append(r, reading)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	tx.Commit()

	if len(r) > 0 && nRows > prevPage.Size {
		cursor = pagination.New(r[0].EpistleID, r[len(r)-1].EpistleID, prevPage.Size)
	}
	return r, cursor, nil
}

const (
	fetchReadingSQL = "SELECT r.status, r.started, r.finished, r.archived, r.created, r.modified, e.link, e.title, e.description, e.favicon, e.created, e.modified FROM reading r JOIN epistles e ON r.epistle_id=e.id WHERE r.epistle_id=$1 AND r.user_id=$2"
)

func Fetch(ctx context.Context, epistleID, userID int64) (reading *Reading, err error) {
	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
		return nil, err
	}
	defer tx.Rollback()

	reading = &Reading{
		EpistleID: epistleID,
		UserID:    userID,
	}
	epistle := &Epistle{
		ID: epistleID,
	}
	if err = tx.QueryRow(fetchReadingSQL, epistleID, userID).Scan(
		&reading.Status,
		&reading.Started,
		&reading.Finished,
		&reading.Archived,
		&reading.Created,
		&reading.Modified,
		&epistle.Link,
		&epistle.Title,
		&epistle.Description,
		&epistle.Favicon,
		&epistle.Created,
		&epistle.Modified); err != nil {
		return nil, err
	}

	reading.epistle = epistle
	return reading, nil
}

const (
	readingUserSQL = "SELECT full_name, email, username, role_id, last_seen, created, modified FROM users WHERE id=$1"
)

// User returns the user associated with the reading. If the user is not cached on the
// struct then a database query is performed and an error may be returned. Use the reset
// bool to force a database query even if the user is cached on the struct.
func (r *Reading) User(ctx context.Context, reset bool) (_ *users.User, err error) {
	if reset || r.user == nil {
		var tx *sql.Tx
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
			return nil, err
		}
		defer tx.Rollback()

		if err = r.fetchUser(tx); err != nil {
			return nil, err
		}

		tx.Commit()
	}
	return r.user, nil
}

func (r *Reading) fetchUser(tx *sql.Tx) (err error) {
	r.user = &users.User{ID: r.UserID}
	if err = tx.QueryRow(readingUserSQL, r.UserID).Scan(&r.user.FullName, &r.user.Email, &r.user.Username, &r.user.RoleID, &r.user.LastSeen, &r.user.Created, &r.user.Modified); err != nil {
		return err
	}
	return nil
}

// Epistle returns the epistle associated with the reading. If the epistle is not cached
// on the struct then a database query is performed and an error may be returned. Use
// the reset bool to force a database query even if the epistle is cached on the struct.
func (r *Reading) Epistle(ctx context.Context, reset bool) (_ *Epistle, err error) {
	if reset || r.epistle == nil {
		var tx *sql.Tx
		if tx, err = db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}); err != nil {
			return nil, err
		}
		defer tx.Rollback()

		r.epistle = &Epistle{ID: r.EpistleID}
		if err = r.epistle.fetch(tx); err != nil {
			return nil, err
		}

		tx.Commit()
	}
	return r.epistle, nil
}
