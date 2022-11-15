package epistles

import (
	"database/sql"
	"errors"
	"time"
)

// Database model for an Epistle object
type Epistle struct {
	ID          int64
	Link        string
	Title       sql.NullString
	Description sql.NullString
	Favicon     sql.NullString
	Created     time.Time
	Modified    time.Time
}

const (
	epistleByLinkSQL = "SELECT id, title, description, favicon, created, modified FROM epistles WHERE link=$1"
	createEpistleSQL = "INSERT INTO epistles (link) VALUES ($1) RETURNING ID"
	epistleTSSQL     = "SELECT created, modified FROM epistles WHERE id=$1"
)

// Get or create an epistle via a URL, which should be unique.
func getOrCreateEpistle(tx *sql.Tx, link string) (e *Epistle, err error) {
	e = &Epistle{Link: link}
	if err = tx.QueryRow(epistleByLinkSQL, link).Scan(&e.ID, &e.Title, &e.Description, &e.Favicon, &e.Created, &e.Modified); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err = tx.QueryRow(createEpistleSQL, link).Scan(&e.ID); err != nil {
				return nil, err
			}

			if err = tx.QueryRow(epistleTSSQL, e.ID).Scan(&e.Created, &e.Modified); err != nil {
				return nil, err
			}

			return e, nil
		}
		return nil, err
	}
	return e, nil
}

const (
	getEpistleSQL = "SELECT link, title, description, favicon, created, modified FROM epistles WHERE id=$1"
)

func (e *Epistle) fetch(tx *sql.Tx) error {
	if e.ID < 1 {
		return ErrIDRequired
	}

	if err := tx.QueryRow(getEpistleSQL, e.ID).Scan(&e.Link, &e.Title, &e.Description, &e.Favicon, &e.Created, &e.Modified); err != nil {
		return err
	}
	return nil
}
