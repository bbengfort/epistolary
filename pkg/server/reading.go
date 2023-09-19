package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server/epistles"
	"github.com/bbengfort/epistolary/pkg/utils/pagination"
	"github.com/bbengfort/epistolary/pkg/utils/sentry"
	"github.com/gin-gonic/gin"
)

func (s *Server) ListReadings(c *gin.Context) {
	var (
		err      error
		out      *api.ReadingPage
		curPage  *pagination.Cursor
		nextPage *pagination.Cursor
	)

	query := &api.PageQuery{}
	if err = c.BindQuery(&query); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse page query"))
		return
	}

	var userID int64
	if userID, err = GetUserID(c); err != nil {
		sentry.Error(c).Err(err).Msg("could not parse userID from request")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	// Parse the previous page token if one was supplied
	if query.PageToken != "" {
		if curPage, err = pagination.Parse(query.PageToken); err != nil {
			sentry.Warn(c).Err(err).Msg("invalid next page token")
			c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
			return
		}
	}

	// Fetch the readings for the user
	var reads []*epistles.Reading
	if reads, nextPage, err = epistles.List(c.Request.Context(), userID, curPage); err != nil {
		sentry.Error(c).Err(err).Msg("could not fetch readings from database")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not fetch readings"))
		return
	}

	out = &api.ReadingPage{
		Readings: make([]*api.Reading, 0, len(reads)),
	}

	if nextPage != nil {
		if out.NextPageToken, err = nextPage.PageToken(); err != nil {
			sentry.Error(c).Err(err).Msg("could not create next page token")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not fetch readings"))
			return
		}
	}

	if curPage != nil {
		prevPage := curPage.PrevPage()
		if out.PrevPageToken, err = prevPage.PageToken(); err != nil {
			sentry.Error(c).Err(err).Msg("could not create prev page token")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not fetch readings"))
			return
		}
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
			Started:     api.Timestamp{Time: r.Started.Time},
			Finished:    api.Timestamp{Time: r.Finished.Time},
			Archived:    api.Timestamp{Time: r.Archived.Time},
			Created:     api.Timestamp{Time: r.Created},
			Modified:    api.Timestamp{Time: r.Modified},
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

	reading.Link = strings.TrimSpace(reading.Link)
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
		sentry.Error(c).Err(err).Msg("could not parse user id")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	if read, err = epistles.Create(c.Request.Context(), userID, reading.Link); err != nil {
		if errors.Is(err, epistles.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, api.ErrorResponse(err))
			return
		}

		sentry.Error(c).Err(err).Msg("could not create reading in database")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	epistle, _ = read.Epistle(c.Request.Context(), false)
	if err = epistle.Sync(c.Request.Context()); err != nil {
		sentry.Error(c).Err(err).Msg("could not sync epistle")
	}

	reading.ID = read.EpistleID
	reading.Status = string(read.Status)
	reading.Link = epistle.Link
	reading.Title = epistle.Title.String
	reading.Description = epistle.Description.String
	reading.Favicon = epistle.Favicon.String
	reading.Started = api.Timestamp{Time: read.Started.Time}
	reading.Finished = api.Timestamp{Time: read.Finished.Time}
	reading.Archived = api.Timestamp{Time: read.Archived.Time}
	reading.Created = api.Timestamp{Time: read.Created}
	reading.Modified = api.Timestamp{Time: read.Modified}

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
		sentry.Error(c).Err(err).Msg("could not parse user id")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	var item *epistles.Reading
	if item, err = epistles.Fetch(c.Request.Context(), reading.ID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, api.ErrorResponse("reading not found"))
			return
		}

		sentry.Error(c).Err(err).Msg("could not fetch readings from database")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
		return
	}

	epistle, _ := item.Epistle(c.Request.Context(), false)
	reading.Status = string(item.Status)
	reading.Link = epistle.Link
	reading.Title = epistle.Title.String
	reading.Description = epistle.Description.String
	reading.Favicon = epistle.Favicon.String
	reading.Started = api.Timestamp{Time: item.Started.Time}
	reading.Finished = api.Timestamp{Time: item.Finished.Time}
	reading.Archived = api.Timestamp{Time: item.Archived.Time}
	reading.Created = api.Timestamp{Time: item.Created}
	reading.Modified = api.Timestamp{Time: item.Modified}

	c.JSON(http.StatusOK, reading)
}

func (s *Server) UpdateReading(c *gin.Context) {
	var (
		err       error
		reading   *api.Reading
		readingID int64
		userID    int64
	)

	// Get the userID from the claims in the request
	if userID, err = GetUserID(c); err != nil {
		sentry.Error(c).Err(err).Msg("could not parse user id")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not process request"))
	}

	// Parse the reading update input from the request
	reading = &api.Reading{}
	if err = c.BindJSON(reading); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse reading input"))
		return
	}

	if readingID, err = strconv.ParseInt(c.Param("readingID"), 10, 64); err != nil {
		c.Error(err)
		c.JSON(http.StatusNotFound, api.ErrorResponse("reading not found"))
		return
	}

	// Populate the reading ID from the endpoint if it was not submitted
	if reading.ID == 0 {
		reading.ID = readingID
	}

	// Ensure the endpoint matches the ID specified by the reading
	if reading.ID != readingID {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("id must match endpoint"))
		return
	}

	reading.Title = strings.TrimSpace(reading.Title)
	reading.Description = strings.TrimSpace(reading.Description)

	// The title is required on PUT requests (otherwise will be unclickable)
	if reading.Title == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("title is required on update"))
		return
	}

	// Convert the reading to the model with the fields that are updateable.
	model := &epistles.Reading{
		EpistleID: reading.ID,
		UserID:    userID,
		Started:   reading.Started.ToSQL(),
		Finished:  reading.Finished.ToSQL(),
		Archived:  reading.Archived.ToSQL(),
	}

	// Convert the reading to the epistle with the fields that are updateable.
	// NOTE: the link of the epistle cannot be updated through this RPC.
	epistle := &epistles.Epistle{
		ID:          reading.ID,
		Title:       sql.NullString{String: reading.Title, Valid: reading.Title != ""},
		Description: sql.NullString{String: reading.Description, Valid: reading.Description != ""},
	}

	if err = epistles.Update(c.Request.Context(), model, epistle); err != nil {
		sentry.Error(c).Err(err).Msg("could not update reading in database")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse(err))
		return
	}

	// Convert the model back to the API for the response
	reading.Status = string(model.Status)
	reading.Started = api.Timestamp{Time: model.Started.Time}
	reading.Finished = api.Timestamp{Time: model.Finished.Time}
	reading.Archived = api.Timestamp{Time: model.Archived.Time}
	reading.Created = api.Timestamp{Time: model.Created}
	reading.Modified = api.Timestamp{Time: model.Modified}

	// Return the reading object back to the user
	c.JSON(http.StatusOK, reading)
}

func (s *Server) DeleteReading(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
