package server

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server/passwd"
	"github.com/bbengfort/epistolary/pkg/server/tokens"
	"github.com/bbengfort/epistolary/pkg/server/users"
	"github.com/bbengfort/epistolary/pkg/utils/sentry"
	sentrylib "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
		sentry.Error(c).Err(err).Msg("could not create derived key")
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("could not register user"))
		return
	}

	// Store the user in the database
	// TODO: handle uniqueness constraints (e.g. username already taken; email already registered, etc.)
	if err = user.Create(c.Request.Context()); err != nil {
		sentry.Error(c).Err(err).Msg("could not create user in database")
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
		Name:     user.FullName.String,
		Username: user.Username,
		Email:    user.Email,
	}
	claims.SetSubjectID(user.ID)

	// Role and permissions should already be on the user from the earlier request.
	role, _ := user.Role(c.Request.Context(), false)
	claims.Role = role.Title
	claims.Permissions, _ = user.Permissions(c.Request.Context(), false)

	if out.AccessToken, out.RefreshToken, err = s.tokens.CreateTokens(claims); err != nil {
		sentry.Error(c).Err(err).Msg("could not create access and refresh tokens")
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	// Set the access and refresh tokens as cookies for the front-end
	if err := SetAuthCookies(c, out.AccessToken, out.RefreshToken, s.conf.Token.CookieDomain); err != nil {
		sentry.Error(c).Err(err).Msg("could not set access and refresh token cookies")
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}

	// Update the last_seen timestamp for the user
	if err = user.UpdateLastSeen(c.Request.Context()); err != nil {
		sentry.Error(c).Err(err).Msg("could not update user last seen")
		c.JSON(http.StatusUnauthorized, api.ErrorResponse("authentication failed"))
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) Logout(c *gin.Context) {
	// Clear cookies but setting the access and refresh tokens to having expired.
	c.SetCookie(AccessTokenCookie, "", -1, "/", s.conf.Token.CookieDomain, true, true)
	c.SetCookie(RefreshTokenCookie, "", -1, "/", s.conf.Token.CookieDomain, true, true)
	c.Status(http.StatusNoContent)
}

const (
	authorization      = "Authorization"
	UserClaims         = "user_claims"
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

// used to extract the access token from the header
var (
	bearer = regexp.MustCompile(`^\s*[Bb]earer\s+([a-zA-Z0-9_\-\.]+)\s*$`)
)

func (s *Server) Authenticate(c *gin.Context) {
	var (
		err    error
		ats    string
		claims *tokens.Claims
	)

	// Parse and verify JWT token in authorization header.
	if ats, err = GetAccessToken(c); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse("authentication required"))
		return
	}

	if claims, err = s.tokens.Verify(ats); err != nil {
		// If the access token is no longer valid, attempt to verify with the refresh token
		var rerr error
		if claims, rerr = s.Reauthenticate(c); rerr != nil {
			c.Error(err)
			c.Error(rerr)
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse("authentication required"))
			return
		}
	}

	// Add claims to context for use in downstream processing and continue
	c.Set(UserClaims, claims)

	// Specify user for Sentry if Sentry is configured
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.Scope().SetUser(sentrylib.User{
			ID:        claims.Subject,
			Email:     claims.Email,
			Name:      claims.Name,
			IPAddress: c.ClientIP(),
		})
	}

	c.Next()
}

func (s *Server) Reauthenticate(c *gin.Context) (_ *tokens.Claims, err error) {
	// Collect and verify the refresh token from the request
	var refresh string
	if refresh, err = GetRefreshToken(c); err != nil {
		return nil, err
	}

	var refreshClaims *tokens.Claims
	if refreshClaims, err = s.tokens.Verify(refresh); err != nil {
		return nil, err
	}

	// Fetch the user from the subject ID of the refreshClaims
	var userID int64
	if userID, err = refreshClaims.SubjectID(); err != nil {
		sentry.Error(c).Err(err).Msg("could not parse subject id from refresh claims")
		return nil, err
	}

	var user *users.User
	if user, err = users.UserFromID(c.Request.Context(), userID); err != nil {
		sentry.Error(c).Err(err).Msg("could not retreive user from id on claims")
		return nil, err
	}

	claims := &tokens.Claims{
		Name:     user.FullName.String,
		Username: user.Username,
		Email:    user.Email,
	}
	claims.SetSubjectID(user.ID)

	// Role and permissions should already be on the user from the earlier request.
	role, _ := user.Role(c.Request.Context(), false)
	claims.Role = role.Title
	claims.Permissions, _ = user.Permissions(c.Request.Context(), false)

	var accessToken, refreshToken string
	if accessToken, refreshToken, err = s.tokens.CreateTokens(claims); err != nil {
		sentry.Error(c).Err(err).Msg("could not create access and refresh tokens")
		return nil, err
	}

	// Set the access and refresh tokens as cookies for the front-end
	if err := SetAuthCookies(c, accessToken, refreshToken, s.conf.Token.CookieDomain); err != nil {
		sentry.Error(c).Err(err).Msg("could not set access and refresh token cookies")
		return nil, err
	}

	log.Debug().Int64("userID", userID).Msg("user reauthenticated")
	return claims, nil
}

func (s *Server) Authorize(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims
		value, exists := c.Get(UserClaims)
		if !exists || value == nil {
			log.Debug().Bool("exists", exists).Msg("user claims interface doesn't exist or is nil")
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse("authorization required"))
			return
		}

		claims, ok := value.(*tokens.Claims)
		if !ok {
			log.Debug().Bool("ok", ok).Msg("user claims interface is not a *tokens.Claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse("authorization required"))
			return
		}

		if !claims.HasAllPermissions(permissions...) {
			log.Trace().Strs("permissions", permissions).Strs("claims", claims.Permissions).Msg("does not have required permissions")
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.ErrorResponse("authorization required"))
			return
		}

		c.Next()
	}
}

func GetAccessToken(c *gin.Context) (tks string, err error) {
	// Attempt to get the access token from the header
	if header := c.GetHeader(authorization); header != "" {
		match := bearer.FindStringSubmatch(header)
		if len(match) == 2 {
			return match[1], nil
		}
		return "", errors.New("could not parse Bearer token from Authorization header")
	}

	// Attempt to get the access token from cookies
	var cookie string
	if cookie, err = c.Cookie(AccessTokenCookie); err == nil {
		// If the error is nil, we were able to retrieve the access token cookie
		return cookie, nil
	}

	// Could not find the access token in the request
	return "", errors.New("no access token found in request")
}

func GetRefreshToken(c *gin.Context) (tks string, err error) {
	if tks, err = c.Cookie(RefreshTokenCookie); err != nil {
		return "", errors.New("no refresh token found in request")
	}
	return tks, nil
}

func GetUserClaims(c *gin.Context) (*tokens.Claims, error) {
	// Get claims
	value, exists := c.Get(UserClaims)
	if !exists || value == nil {
		return nil, errors.New("no user claims exist on request")
	}

	claims, ok := value.(*tokens.Claims)
	if !ok {
		return nil, errors.New("incorrect claims type stored on context")
	}
	return claims, nil
}

func GetUserID(c *gin.Context) (int64, error) {
	claims, err := GetUserClaims(c)
	if err != nil {
		return 0, err
	}
	return claims.SubjectID()
}

// SetAuthCookies is a helper function that sets access and refresh token cookies on a
// gin request. The access token cookie (access_token) is an http only cookie that
// expires when the access token expires. The refresh token cookie is an http only
// cookie (it can't be accessed by client-side scripts) and it expires when the refresh
// token expires. Both cookies require https and will not be set (silently) over http.
func SetAuthCookies(c *gin.Context, accessToken, refreshToken, domain string) (err error) {
	var accessExpires time.Time
	if accessExpires, err = tokens.ExpiresAt(accessToken); err != nil {
		return err
	}

	// Set the access token cookie: httpOnly is true; cannot be accessed by javascript
	accessMaxAge := int((time.Until(accessExpires)).Seconds())
	c.SetCookie(AccessTokenCookie, accessToken, accessMaxAge, "/", domain, true, true)

	var refreshExpires time.Time
	if refreshExpires, err = tokens.ExpiresAt(refreshToken); err != nil {
		return err
	}

	// Set the refresh token cookie; httpOnly is true; cannot be accessed by javascript
	refreshMaxAge := int((time.Until(refreshExpires)).Seconds())
	c.SetCookie(RefreshTokenCookie, refreshToken, refreshMaxAge, "/", domain, true, true)

	return nil
}
