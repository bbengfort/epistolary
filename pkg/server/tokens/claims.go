package tokens

import jwt "github.com/golang-jwt/jwt/v4"

// Claims implements custom claims for the Epistolary application.
type Claims struct {
	jwt.RegisteredClaims
	Name        string   `json:"name,omitempty"`
	Username    string   `json:"username,omitempty"`
	Email       string   `json:"email,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}
