// Package runner provides models and methods to interact with browsers
// to perform HTTP requests on websites. It is used to generate load and
// to delivers insights on the implication of that load.
package runner

import (
	"context"
	"errors"
)

var (
	// ErrInvalidContext is an error indicating that a given context is no runner context.
	ErrInvalidContext = errors.New("not a runner context")

	// ErrNoNetworkEventFound is an error indicating that no network event was found.
	ErrNoNetworkEventFound = errors.New("no network event for base url found")
)

type contextKey struct{}

// Runner is an abstraction to represent objects interacting with browsers.
// Implementations of runner's are able to communicate with browsers (or other HTTP capable libraries).
type Runner interface {
	WithContext(ctx context.Context) context.Context
}

// FromContext extracts the runner instance from ctx.
func FromContext(ctx context.Context) Runner {
	v, _ := ctx.Value(contextKey{}).(Runner)
	return v
}
