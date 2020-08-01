package runner

import (
	"context"
)

// FakeRunner is a runner which fakes requests.
// Instead of performing real HTTP requests, it only returns fake values.
// This is useful for testing purposes.
type FakeRunner uint

// NewFakeRunner returns a new FakeRunner.
func NewFakeRunner(id int) *FakeRunner {
	f := FakeRunner(id)
	return &f
}

// WithContext returns a new runner context derived from ctx.
func (r *FakeRunner) WithContext(ctx context.Context) context.Context {
	runnerCtx := context.WithValue(ctx, contextKey{}, r)
	return runnerCtx
}
