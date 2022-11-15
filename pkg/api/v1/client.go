package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

// New creates a new api.v1 API client that implements the EpistolaryClient interface.
func New(endpoint string) (_ EpistolaryClient, err error) {
	c := &APIv1{
		client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Timeout:       30 * time.Second,
		},
	}

	// Create cookie jar
	if c.client.Jar, err = cookiejar.New(nil); err != nil {
		return nil, fmt.Errorf("could not create cookiejar: %s", err)
	}

	if c.endpoint, err = url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("could not parse endpoint: %s", err)
	}
	return c, nil
}

// APIv1 implements the EpistolaryClient interface.
type APIv1 struct {
	endpoint *url.URL
	client   *http.Client
}

// Ensure the API implments the EpistolaryClient interface.
var _ EpistolaryClient = &APIv1{}

//===========================================================================
// Client Methods
//===========================================================================

func (s *APIv1) Register(ctx context.Context, in *RegisterRequest) (err error) {
	//  Make the HTTP request
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodPost, "/v1/register", in, nil); err != nil {
		return err
	}

	if _, err = s.Do(req, nil, true); err != nil {
		return err
	}

	return nil
}

func (s *APIv1) Login(ctx context.Context, in *LoginRequest) (out *LoginReply, err error) {
	//  Make the HTTP request
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodPost, "/v1/login", in, nil); err != nil {
		return nil, err
	}

	out = &LoginReply{}
	if _, err = s.Do(req, out, true); err != nil {
		return nil, err
	}

	// TODO: save access and refresh token for follow up request handling
	return out, nil
}

func (s *APIv1) ListReadings(ctx context.Context, in *PageQuery) (out *ReadingPage, err error) {
	var params url.Values
	if params, err = query.Values(in); err != nil {
		return nil, fmt.Errorf("could not encode query params: %w", err)
	}

	//  Make the HTTP request
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodGet, "/v1/reading", nil, &params); err != nil {
		return nil, err
	}

	out = &ReadingPage{}
	if _, err = s.Do(req, out, true); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *APIv1) CreateReading(ctx context.Context, in *Reading) (out *Reading, err error) {
	//  Make the HTTP request
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodPost, "/v1/reading", in, nil); err != nil {
		return nil, err
	}

	out = &Reading{}
	if _, err = s.Do(req, out, true); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *APIv1) FetchReading(ctx context.Context, id int64) (out *Reading, err error) {
	//  Make the HTTP request
	endpoint := fmt.Sprintf("/v1/reading/%d", id)
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return nil, err
	}

	out = &Reading{}
	if _, err = s.Do(req, out, true); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *APIv1) UpdateReading(ctx context.Context, in *Reading) (out *Reading, err error) {
	//  Make the HTTP request
	endpoint := fmt.Sprintf("/v1/reading/%d", in.ID)
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodPut, endpoint, in, nil); err != nil {
		return nil, err
	}

	out = &Reading{}
	if _, err = s.Do(req, out, true); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *APIv1) DeleteReading(ctx context.Context, id int64) (err error) {
	//  Make the HTTP request
	endpoint := fmt.Sprintf("/v1/reading/%d", id)
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodDelete, endpoint, nil, nil); err != nil {
		return err
	}

	if _, err = s.Do(req, nil, true); err != nil {
		return err
	}
	return nil
}

func (s *APIv1) Status(ctx context.Context) (out *StatusReply, err error) {
	//  Make the HTTP request
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodGet, "/v1/status", nil, nil); err != nil {
		return nil, err
	}

	// Execute the request and get a response
	// NOTE: cannot use s.Do because we want to parse 503 Unavailable errors
	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return nil, fmt.Errorf("could not execute request: %s", err)
	}
	defer rep.Body.Close()

	// Detect other errors
	if rep.StatusCode != http.StatusOK && rep.StatusCode != http.StatusServiceUnavailable {
		return nil, fmt.Errorf("[%d] %s", rep.StatusCode, rep.Status)
	}

	// Deserialize the JSON data from the response
	out = &StatusReply{}
	if err = json.NewDecoder(rep.Body).Decode(out); err != nil {
		return nil, fmt.Errorf("could not deserialize StatusReply: %s", err)
	}
	return out, nil
}

//===========================================================================
// Helper Methods
//===========================================================================

const (
	userAgent    = "Epistolary API Client/v1"
	accept       = "application/json"
	acceptLang   = "en-US,en"
	acceptEncode = "gzip, deflate, br"
	contentType  = "application/json; charset=utf-8"
)

// NewRequest creates an http.Request with the specified context and method, resolving
// the path to the root endpoint of the API (e.g. /v1) and serializes the data to JSON.
// This method also sets the default headers of all Persona v1 client requests.
func (s *APIv1) NewRequest(ctx context.Context, method, path string, data interface{}, params *url.Values) (req *http.Request, err error) {
	// Resolve the URL reference from the path
	endpoint := s.endpoint.ResolveReference(&url.URL{Path: path})
	if params != nil && len(*params) > 0 {
		endpoint.RawQuery = params.Encode()
	}

	var body io.ReadWriter
	if data != nil {
		body = &bytes.Buffer{}
		if err = json.NewEncoder(body).Encode(data); err != nil {
			return nil, fmt.Errorf("could not serialize request data: %s", err)
		}
	} else {
		body = nil
	}

	// Create the http request
	if req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), body); err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}

	// Set the headers on the request
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", accept)
	req.Header.Add("Accept-Language", acceptLang)
	req.Header.Add("Accept-Encoding", acceptEncode)
	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// Do executes an http request against the server, performs error checking, and
// deserializes the response data into the specified struct if requested.
func (s *APIv1) Do(req *http.Request, data interface{}, checkStatus bool) (rep *http.Response, err error) {
	if rep, err = s.client.Do(req); err != nil {
		return rep, fmt.Errorf("could not execute request: %s", err)
	}
	defer rep.Body.Close()

	// Detect errors if they've occurred
	if checkStatus {
		if rep.StatusCode < 200 || rep.StatusCode >= 300 {
			// Attempt to read the error response from the JSON, ignore body
			// deserialization or read errors and simply return the status error.
			var reply Reply
			if err = json.NewDecoder(rep.Body).Decode(&reply); err == nil {
				if reply.Error != "" {
					return rep, fmt.Errorf("[%d] %s", rep.StatusCode, reply.Error)
				}
			}
			return rep, errors.New(rep.Status)
		}
	}

	// Check the content type to ensure data deserialization is possible
	if ct := rep.Header.Get("Content-Type"); ct != contentType {
		return rep, fmt.Errorf("unexpected content type: %q", ct)
	}

	// Deserialize the JSON data from the body
	if data != nil && rep.StatusCode >= 200 && rep.StatusCode < 300 {
		if err = json.NewDecoder(rep.Body).Decode(data); err != nil {
			return nil, fmt.Errorf("could not deserialize response data: %s", err)
		}
	}

	return rep, nil
}
