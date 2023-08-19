package epistles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bbengfort/epistolary/pkg/server/db"
	"github.com/bbengfort/epistolary/pkg/server/fetch"
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

func (e *Epistle) IsSynced() bool {
	return e.Title.String != "" || e.Description.String != "" || e.Favicon.String != ""
}

func (e *Epistle) Sync(ctx context.Context) (err error) {
	if e.Link == "" {
		return ErrLinkRequired
	}

	var doc *fetch.Document
	if doc, err = fetch.Fetch(ctx, e.Link); err != nil {
		return err
	}

	if doc.FaviconCheck, err = fetch.CheckIcon(ctx, doc.Favicon); err != nil {
		return err
	}

	if !doc.FaviconCheck {
		doc.Favicon = ""
	}

	e.Title = sql.NullString{Valid: doc.Title != "", String: doc.Title}
	e.Description = sql.NullString{Valid: doc.Description != "", String: doc.Description}
	e.Favicon = sql.NullString{Valid: doc.Favicon != "", String: doc.Favicon}

	// Save the epistle after fetching it
	return e.Save(ctx)
}

const (
	saveEpistleSQL = "UPDATE epistles SET link=$2, title=%3, description=$4, favicon=$5, created=$6, modified=$7 WHERE id=$1"
)

func (e *Epistle) Save(ctx context.Context) (err error) {
	var tx *sql.Tx
	if tx, err = db.BeginTx(ctx, nil); err != nil {
		return fmt.Errorf("could not start write tx: %w", err)
	}
	defer tx.Rollback()

	if err = e.save(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (e *Epistle) save(tx *sql.Tx) (err error) {
	if e.ID == 0 {
		return ErrIDRequired
	}

	e.Modified = time.Now()
	if _, err = tx.Exec(saveEpistleSQL, e.ID, e.Link, e.Title, e.Description, e.Favicon, e.Created, e.Modified); err != nil {
		return fmt.Errorf("could not save epistle: %w", err)
	}

	return nil
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
