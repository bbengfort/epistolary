package api

import (
	"context"
)

//===========================================================================
// Service Interface
//===========================================================================

type EpistolaryClient interface {
	Status(ctx context.Context) (*StatusReply, error)
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
