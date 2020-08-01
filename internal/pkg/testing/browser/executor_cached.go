package browser

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

// CachedTestExecutor implements the Executor interface and can be used in tests as a
// testify mock object. It simulates a cached network event while listening for a devtools target event.
type CachedTestExecutor struct {
	mock.Mock
}

// NewCachedTestExecutor returns a new CachedTestExecutor.
func NewCachedTestExecutor() *CachedTestExecutor {
	return &CachedTestExecutor{}
}

// Run registers the method parameters and returns an error value declared by a testify mock setup.
// This can be used to assert a correct method call.
func (e *CachedTestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

// ListenTarget registers the method parameters, which can be asserted in unit tets.
// It also generates a network event object containing a cached result to test correct event listening behaviour.
func (e *CachedTestExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
	e.Called(ctx, fn)

	// Simulate a network event
	ev := &network.EventResponseReceived{
		RequestID: "testing",
		Type:      network.ResourceTypeDocument,
		Response: &network.Response{
			URL:        URL,
			Status:     Status,
			StatusText: StatusText,
			Timing: &network.ResourceTiming{
				ConnectStart:      -1,
				ReceiveHeadersEnd: ReceiveHeadersEnd,
			},
		},
	}

	fn(ev)
}
