package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

// Executor abstracts away the implementation details of communicating to a Chrome Devtools protocol
// capable library.
type Executor interface {
	Run(ctx context.Context, actions ...chromedp.Action) error
	ListenTarget(ctx context.Context, fn func(ev interface{}))
}

// ChromeDPExecutor implements the Executor library as a chromedp wrapper.
type ChromeDPExecutor struct{}

// New returns a new ChromeDPExecutor.
func New() *ChromeDPExecutor {
	return &ChromeDPExecutor{}
}

// Run executes a chromedp Run method, performing the actual request on a browser.
// See chromedp documentation for further details.
func (e *ChromeDPExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	return chromedp.Run(ctx, actions...)
}

// ListenTarget registers a new function as a listener, which is triggered on incoming CDP events.
// See chromedp documentation for further details.
func (e *ChromeDPExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
	chromedp.ListenTarget(ctx, fn)
}
