package fetch

import (
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
)

func Fetch(ctx context.Context, url string) (*Document, error) {
	fetcher := NewHTMLFetcher(url)
	return fetcher.Fetch(ctx)
}

// HTMLFetcher is an interface for fetching the full HTML associated with a feed item
type HTMLFetcher struct {
	url string // the url of the article full text
}

// NewHTMLFetcher creates a new HTML fetcher that can fetch the full HTML from the specified URL.
func NewHTMLFetcher(url string) *HTMLFetcher {
	return &HTMLFetcher{
		url: url,
	}
}

// The HTMLFetcher uses GET requests to retrieve details about the HTML page.
func (f *HTMLFetcher) Fetch(ctx context.Context) (doc *Document, err error) {
	var req *http.Request
	if req, err = f.newRequest(ctx); err != nil {
		return nil, err
	}

	var rep *http.Response
	if rep, err = client.Do(req); err != nil {
		return nil, err
	}

	// Close the body of the response reader when we're done.
	if rep != nil && rep.Body != nil {
		defer rep.Body.Close()
	}

	// Check the status code of the response; note that 304 means not modified, but we
	// are still returning a 304 error to signal to the Subscription that nothing has
	// changed and that the post is nil.
	if rep.StatusCode < 200 || rep.StatusCode >= 300 {
		return nil, HTTPError{
			Status: rep.Status,
			Code:   rep.StatusCode,
		}
	}

	// Handle compression
	// TODO: handle deflate, br, and other compression schemes
	var reader io.Reader
	switch encoding := rep.Header.Get(HeaderContentEncoding); encoding {
	case gzipEncode:
		var gzread *gzip.Reader
		if gzread, err = gzip.NewReader(rep.Body); err != nil {
			return nil, err
		}
		defer gzread.Close()
		reader = gzread
	case brotliEncode:
		reader = brotli.NewReader(rep.Body)
	case lzwEncode:
		// TODO: what should the order and litwidth be?
		lzwreader := lzw.NewReader(rep.Body, lzw.MSB, 8)
		defer lzwreader.Close()
		reader = lzwreader
	case zlibEncode:
		var zlibreader io.ReadCloser
		if zlibreader, err = zlib.NewReader(rep.Body); err != nil {
			return nil, err
		}
		defer zlibreader.Close()
		reader = zlibreader
	case "":
		reader = rep.Body
	default:
		return nil, fmt.Errorf("unknown content encoding %q", encoding)
	}

	var tree *goquery.Document
	if tree, err = goquery.NewDocumentFromReader(reader); err != nil {
		return nil, err
	}

	doc = &Document{
		Link:  req.URL.String(),
		Title: tree.Find("title").Contents().Text(),
	}

	tree.Find("meta").EachWithBreak(func(index int, item *goquery.Selection) bool {
		if item.AttrOr("name", "") == "description" {
			doc.Description = item.AttrOr("content", "")
			if doc.Description != "" {
				return false
			}
		}
		return true
	})

	tree.Find("link").EachWithBreak(func(index int, item *goquery.Selection) bool {
		if item.AttrOr("rel", "") == "icon" || item.AttrOr("rel", "") == "shortcut icon" {
			doc.Favicon = item.AttrOr("href", "")
			if doc.Favicon != "" {
				return false
			}
		}
		return true
	})

	if doc.Favicon == "" {
		// Default to the favicon.ico at the root of the domain.
		doc.Favicon = req.URL.ResolveReference(&url.URL{Path: "/favicon.ico"}).String()
	}
	return doc, nil
}

func (f *HTMLFetcher) newRequest(ctx context.Context) (req *http.Request, err error) {
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil); err != nil {
		return nil, err
	}

	req.Header.Set(HeaderUserAgent, userAgent)
	req.Header.Set(HeaderAccept, acceptHTML)
	req.Header.Set(HeaderAcceptLang, acceptLang)
	req.Header.Set(HeaderAcceptEncode, acceptEncode)
	req.Header.Set(HeaderCacheControl, cacheControl)
	req.Header.Set(HeaderReferer, referer)

	return req, nil
}

type Document struct {
	Link         string `json:"link"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Favicon      string `json:"favicon"`
	FaviconCheck bool   `json:"favicon_exists,omitempty"`
}
