package api

import (
	"context"
	"time"
)

//===========================================================================
// Service Interface
//===========================================================================

type EpistolaryClient interface {
	Register(context.Context, *RegisterRequest) error
	Login(context.Context, *LoginRequest) (*LoginReply, error)
	Status(context.Context) (*StatusReply, error)

	ListReadings(context.Context, *PageQuery) (*ReadingPage, error)
	CreateReading(context.Context, *Reading) (*Reading, error)
	FetchReading(_ context.Context, id int64) (*Reading, error)
	UpdateReading(context.Context, *Reading) (*Reading, error)
	DeleteReading(_ context.Context, id int64) error
}

//===========================================================================
// Top Level Requests and Responses
//===========================================================================

// Reply contains standard fields that are used for generic API responses and errors.
type Reply struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

// StatusReply is returned on status requests. Note that no request is needed.
type StatusReply struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime,omitempty"`
	Version string `json:"version,omitempty"`
}

// PageQuery allows the user to request the next or previous page from a given cursor.
type PageQuery struct {
	PageSize      uint64 `url:"page_size,omitempty" form:"page_size" json:"page_size,omitempty"`
	NextPageToken string `url:"next_page_token,omitempty" form:"next_page_token" json:"next_page_token,omitempty"`
}

//===========================================================================
// Epistolary v1 API Requests and Responses
//===========================================================================

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReply struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ReadingPage struct {
	Readings      []*Reading `json:"readings"`
	NextPageToken string     `json:"next_page_token"`
	PrevPageToken string     `json:"prev_page_token"`
}

type Reading struct {
	ID          int64     `json:"id,omitempty"`
	Status      string    `json:"staus,omitempty"`
	Link        string    `json:"link"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	Favicon     string    `json:"favicon,omitempty"`
	Started     time.Time `json:"started,omitempty"`
	Finished    time.Time `json:"finished,omitempty"`
	Archived    time.Time `json:"archived,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Modified    time.Time `json:"modified,omitempty"`
}

//===========================================================================
// OpenID Configuration
//===========================================================================

type OpenIDConfiguration struct {
	Issuer                        string   `json:"issuer"`
	AuthorizationEP               string   `json:"authorization_endpoint"`
	TokenEP                       string   `json:"token_endpoint"`
	DeviceAuthorizationEP         string   `json:"device_authorization_endpoint"`
	UserInfoEP                    string   `json:"userinfo_endpoint"`
	MFAChallengeEP                string   `json:"mfa_challenge_endpoint"`
	JWKSURI                       string   `json:"jwks_uri"`
	RegistrationEP                string   `json:"registration_endpoint"`
	RevocationEP                  string   `json:"revocation_endpoint"`
	ScopesSupported               []string `json:"scopes_supported"`
	ResponseTypesSupported        []string `json:"response_types_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	ResponseModesSupported        []string `json:"response_modes_supported"`
	SubjectTypesSupported         []string `json:"subject_types_supported"`
	IDTokenSigningAlgValues       []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethods      []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported               []string `json:"claims_supported"`
	RequestURIParameterSupported  bool     `json:"request_uri_parameter_supported"`
}
