package databackend

import "time"

// DataBackend is an abstract type for handling persistence
// of result data.
type DataBackend interface {
	// Store stores a result.
	Store(*Result) error

	// Close closes all connections or handles to it's backend
	// and should be called, when no more results need to persisted.
	Close() error
}

// Result represents the result of one call to an endpoint.
type Result struct {
	URL               string        `json:"url"`
	HttpStatusCode    int           `json:"httpstatuscode"`
	HttpStatusMessage string        `json:"httpstatusmessage"`
	Ttfb              time.Duration `json:"ttfb"`
	Cached            bool          `json:"cached"`
}
