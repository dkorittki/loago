// Package browser contains mocked Executor implementations for the use in unit tests.
package browser

import (
	"context"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

// TestExecutor implements the Executor interface and can be used in tests as a
// testify mock object.
type TestExecutor struct {
	mock.Mock
}

// NewTestExecutor returns a new TestExecutor.
func NewTestExecutor() *TestExecutor {
	return &TestExecutor{}
}

// Run registers the method parameters and returns an error value declared by a testify mock setup.
// This can be used to assert a correct method call.
func (e *TestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

// ListenTarget registers the method parameters, which can be asserted in unit tets.
func (e *TestExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
	e.Called(ctx, fn)
}
