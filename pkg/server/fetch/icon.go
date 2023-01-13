package fetch

import (
	"context"
	"net/http"
	"strings"
)

// CheckIcon returns true if the icon at the specified URL can be retrieved and has an
// image mimetype, otherwise it returns false.
func CheckIcon(ctx context.Context, url string) (bool, error) {
	fetcher := NewIconFetcher(url)
	return fetcher.Check(ctx)
}

type IconFetcher struct {
	HTMLFetcher
}

func NewIconFetcher(url string) *IconFetcher {
	return &IconFetcher{
		HTMLFetcher: HTMLFetcher{
			url: url,
		},
	}
}

func (f *IconFetcher) Check(ctx context.Context) (exists bool, err error) {
	var req *http.Request
	if req, err = f.newRequest(ctx); err != nil {
		return false, err
	}

	var rep *http.Response
	if rep, err = client.Do(req); err != nil {
		return false, err
	}

	// Close the body of the response reader when we're done.
	if rep != nil && rep.Body != nil {
		defer rep.Body.Close()
	}

	// Check the status code of the response, if not a 200 then the icon does not exist
	if rep.StatusCode < 200 || rep.StatusCode >= 300 {
		return false, nil
	}

	// Ensure the content type starts with image/
	ctype := rep.Header.Get(HeaderContentType)
	ctype = strings.ToLower(strings.TrimSpace(ctype))
	if !strings.HasPrefix(ctype, "image/") {
		return false, nil
	}

	// All of our checks pass so we assume that the icon is correct.
	return true, nil
}
