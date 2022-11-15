package tokens

import (
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
)

// Claims implements custom claims for the Epistolary application.
type Claims struct {
	jwt.RegisteredClaims
	Name        string   `json:"name,omitempty"`
	Username    string   `json:"username,omitempty"`
	Email       string   `json:"email,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func (c *Claims) SetSubjectID(uid int64) {
	c.Subject = strconv.FormatInt(uid, 16)
}

func (c Claims) SubjectID() (int64, error) {
	return strconv.ParseInt(c.Subject, 16, 64)
}

func (c Claims) HasPermission(required string) bool {
	for _, permisison := range c.Permissions {
		if permisison == required {
			return true
		}
	}
	return false
}

func (c Claims) HasAllPermissions(required ...string) bool {
	for _, perm := range required {
		if !c.HasPermission(perm) {
			return false
		}
	}
	return true
}
