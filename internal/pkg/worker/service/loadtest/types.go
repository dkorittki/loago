package loadtest

import (
	"time"
)

// An Endpoint represents a URL and a weight indicating the "importance" of the URL.
type Endpoint struct {
	// The URL on which the request will be performed.
	URL string

	// The "importance" of the URL. The higher the number,
	// the more often a request on the endpoint will be made.
	Weight uint
}

// EndpointResult contains all necessary information of a runners response results.
type EndpointResult struct {
	// URL is the ressource requested by the runner.
	URL string

	// HTTPStatusCode is the http status code of the runners response.
	HTTPStatusCode int

	// HTTPStatusMessage is the http status message of the runners response.
	HTTPStatusMessage string

	// TTFB is the time-to-first-byte of the runners response.
	TTFB time.Duration

	// Cached indicates if the browser cache was used instead of performing a real request.
	Cached bool
}

// BrowserType represents a type of browser.
type BrowserType int

const (
	// BrowserTypeFake represents a fake browser type.
	BrowserTypeFake BrowserType = 0

	// BrowserTypeChrome represents a chrome browser type.
	BrowserTypeChrome BrowserType = 1
)
