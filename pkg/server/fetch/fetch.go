package fetch

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Fetcher is an interface for statefully making periodic requests to a resource.
type Fetcher interface {
	Fetch(context.Context) (any, error)
}

// A package level http client for making requests. It is best practice to not use the
// default http.Client but to use your own with timeouts correctly specified. The
// package also admonishes us to only create one client for efficiency because the
// client is itself thread safe.The client is initialized by init() and can be modified
// using the SetDefaultClient function (e.g. for testing). All HTTP based fetchers
// should use this client.
var client *http.Client

func init() {
	// Initialize the HTTP client used in this package.
	jar, _ := cookiejar.New(nil)
	dialer := &net.Dialer{Timeout: 45 * time.Second}
	client = &http.Client{
		Timeout:       1 * time.Minute, // long time out enables global fetch
		CheckRedirect: nil,             // default policy is try following redirect 10 times
		Transport: &http.Transport{
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: 45 * time.Second,
			DisableKeepAlives:   true,
			DisableCompression:  false,
		},
		Jar: jar,
	}
}

// Header values to send along with requests made by the fetch package.
const (
	userAgent    = "Epistolary/v1"
	acceptHTML   = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	acceptRSS    = "application/atom+xml,application/rdf+xml,application/rss+xml,application/x-netcdf,application/xml;q=0.9,text/xml;q=0.2,*/*;q=0.1"
	acceptLang   = "*"
	acceptEncode = "gzip,deflate,br,*"
	referer      = ""
	cacheControl = "max-age=3600"
	aimType      = "feed"
	gzipEncode   = "gzip"
	brotliEncode = "br"
	zlibEncode   = "deflate"
	lzwEncode    = "compress"
)

// Canonical names of headers used by the fetch package
const (
	HeaderUserAgent       = "User-Agent"
	HeaderAccept          = "Accept"
	HeaderAcceptLang      = "Accept-Language"
	HeaderAcceptEncode    = "Accept-Encoding"
	HeaderCacheControl    = "Cache-Control"
	HeaderReferer         = "Referer"
	HeaderIfNoneMatch     = "If-None-Match"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderRFC3229         = "A-IM"
	HeaderETag            = "ETag"
	HeaderLastModified    = "Last-Modified"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentType     = "Content-Type"
)

// SetClient allows you to specify an alternative http.Client to the default one
// used by all http based Fetchers in this package. Use this function to change the
// timeouts of the client or to set a test client.
func SetClient(c *http.Client) {
	client = c
}
