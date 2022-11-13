package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server/passwd"
	"github.com/bbengfort/epistolary/pkg/server/tokens"
	"github.com/bbengfort/epistolary/pkg/server/users"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func (s *Server) Register(c *gin.Context) {
	var in *api.RegisterRequest
	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse register request"))
		return
	}

	// Check required fields
	if in.Email == "" || in.Username == "" || in.Password == "" {
		c.Error(errors.New("missing required field"))
		c.JSON(http.StatusBadRequest, api.ErrorResponse("email, username, and password are required"))
		return
	}

	// Create a user with the given information
	user := &users.User{
		FullName: sql.NullString{},
		Email:    in.Email,
		Username: in.Username,
	}

	if in.FullName != "" {
		user.FullName.String = in.FullName
		user.FullName.Valid = true
	}

	// Create an argon2 derived key for storing the password
	var err error
	if user.Password, err = passwd.CreateDerivedKey(in.Password); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not register user"))
		return
	}

	// Store the user in the database
	// TODO: handle uniqueness constraints (e.g. username already taken; email already registered, etc.)
	if err = user.Create(c.Request.Context()); err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not register user"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *Server) Login(c *gin.Context) {
	var (
		err  error
		in   *api.LoginRequest
		out  *api.LoginReply
		user *users.User
	)

	if err := c.BindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse("could not parse login request"))
		return
	}

	// Check required fields
	if in.Username == "" || in.Password == "" {
		c.Error(errors.New("missing required field"))
		c.JSON(http.StatusBadRequest, api.ErrorResponse("username and password are required"))
		return
	}

	// Fetch the user from the database
	if user, err = users.UserFromUsername(c.Request.Context(), in.Username, true); err != nil {
		c.Error(err)
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	// Verify the derived key for the user
	// NOTE: if the user is not verified we MUST not proceed with the rest of the function!
	if verified, err := passwd.VerifyDerivedKey(user.Password, in.Password); err != nil || !verified {
		if err != nil {
			c.Error(err)
		}
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	// Create the access and refresh tokens from the claims
	out = &api.LoginReply{}
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.FormatInt(user.ID, 16),
		},
		Name:     user.FullName.String,
		Username: user.Username,
		Email:    user.Email,
	}

	// Role and permissions should already be on the user from the earlier request.
	role, _ := user.Role(c.Request.Context(), false)
	claims.Role = role.Title
	claims.Permissions, _ = user.Permissions(c.Request.Context(), false)

	if out.AccessToken, out.RefreshToken, err = s.tokens.CreateTokens(claims); err != nil {
		c.Error(err)
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	// Update the last_seen timestamp for the user
	if err = user.UpdateLastSeen(c.Request.Context()); err != nil {
		c.Error(err)
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	c.JSON(http.StatusOK, out)
}
