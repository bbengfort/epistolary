package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server/epistles"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (s *Server) ListReadings(c *gin.Context) {
	var (
		err error
		out *api.ReadingPage
	)

	// TODO: add pagination
	query := &api.PageQuery{}
	if err = c.BindQuery(&query); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse page query"))
		return
	}

	var userID int64
	if userID, err = GetUserID(c); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	// Fetch the readings for the user
	var reads []*epistles.Reading
	if reads, err = epistles.List(c.Request.Context(), userID); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not fetch readings"))
		return
	}

	out = &api.ReadingPage{
		Readings: make([]*api.Reading, 0, len(reads)),
	}

	for _, r := range reads {
		epistle, _ := r.Epistle(c.Request.Context(), false)
		item := &api.Reading{
			ID:          r.EpistleID,
			Status:      string(r.Status),
			Link:        epistle.Link,
			Title:       epistle.Title.String,
			Description: epistle.Description.String,
			Favicon:     epistle.Favicon.String,
			Started:     r.Started.Time,
			Finished:    r.Finished.Time,
			Archived:    r.Archived.Time,
			Created:     r.Created,
			Modified:    r.Modified,
		}

		out.Readings = append(out.Readings, item)
	}

	c.JSON(http.StatusOK, out)
}

func (s *Server) CreateReading(c *gin.Context) {
	var (
		err     error
		read    *epistles.Reading
		epistle *epistles.Epistle
	)

	reading := &api.Reading{}
	if err = c.BindJSON(&reading); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse reading input"))
		return
	}

	if reading.Link == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("link required to create reading"))
		return
	}

	if reading.ID != 0 || reading.Title != "" || reading.Description != "" || reading.Favicon != "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("reading can only be created with a link"))
		return
	}

	var userID int64
	if userID, err = GetUserID(c); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	if read, err = epistles.Create(c.Request.Context(), userID, reading.Link); err != nil {
		// TODO: handle constraint violations better
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	epistle, _ = read.Epistle(c.Request.Context(), false)
	if err = epistle.Sync(c.Request.Context()); err != nil {
		log.Error().Err(err).Msg("could not sync epistle")
	}

	reading.ID = read.EpistleID
	reading.Status = string(read.Status)
	reading.Link = epistle.Link
	reading.Title = epistle.Title.String
	reading.Description = epistle.Description.String
	reading.Favicon = epistle.Favicon.String
	reading.Started = read.Started.Time
	reading.Finished = read.Finished.Time
	reading.Archived = read.Archived.Time
	reading.Created = read.Created
	reading.Modified = read.Modified

	c.JSON(http.StatusCreated, reading)
}

func (s *Server) FetchReading(c *gin.Context) {
	var err error
	reading := &api.Reading{}
	readingID := c.Param("readingID")

	if reading.ID, err = strconv.ParseInt(readingID, 10, 64); err != nil {
		c.Error(err)
		c.JSON(http.StatusNotFound, api.ErrorResponse("reading not found"))
		return
	}

	var userID int64
	if userID, err = GetUserID(c); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	var item *epistles.Reading
	if item, err = epistles.Fetch(c.Request.Context(), reading.ID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, api.ErrorResponse("reading not found"))
			return
		}

		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	epistle, _ := item.Epistle(c.Request.Context(), false)
	reading.Status = string(item.Status)
	reading.Link = epistle.Link
	reading.Title = epistle.Title.String
	reading.Description = epistle.Description.String
	reading.Favicon = epistle.Favicon.String
	reading.Started = item.Started.Time
	reading.Finished = item.Finished.Time
	reading.Archived = item.Archived.Time
	reading.Created = item.Created
	reading.Modified = item.Modified

	c.JSON(http.StatusOK, reading)
}

func (s *Server) UpdateReading(c *gin.Context) {}

func (s *Server) DeleteReading(c *gin.Context) {}
