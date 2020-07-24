package loadtest

import (
	"time"
)

type Endpoint struct {
	URL    string
	Weight uint
}

type EndpointResult struct {
	URL               string
	HTTPStatusCode    int
	HTTPStatusMessage string
	TTFB              time.Duration
	Cached            bool
}

type BrowserType int

const (
	BrowserTypeFake   BrowserType = 0
	BrowserTypeChrome BrowserType = 1
)
