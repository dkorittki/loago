package runner

import (
	"context"
	"errors"
)

var (
	InvalidContextError      = errors.New("not a runner context")
	NoNetworkEventFoundError = errors.New("no network event for base url found")
)

type contextKey struct{}

type Runner interface {
	WithContext(ctx context.Context) context.Context
}

// FromContext extracts the runner instance from ctx.
func FromContext(ctx context.Context) Runner {
	v, _ := ctx.Value(contextKey{}).(Runner)
	return v
}
