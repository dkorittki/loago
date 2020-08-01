package browser

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

var (
	// URL which is used in the simulated network event.
	URL = "http://foo.bar"

	// Status used in the simulated network event.
	Status = int64(200)

	// StatusText used in the simulated network event.
	StatusText = "Testing OK"

	// ConnectStart value used in the simulated network event.
	ConnectStart = float64(60)

	// ReceiveHeadersEnd value used in the simulated network event.
	ReceiveHeadersEnd = float64(100)
)

// EventTestExecutor implements the Executor interface and can be used in tests as a
// testify mock object. It simulates a network event while listening for a devtools target event.
type EventTestExecutor struct {
	mock.Mock
}

// NewEventTestExecutor returns a new EventTestExecutor.
func NewEventTestExecutor() *EventTestExecutor {
	return &EventTestExecutor{}
}

// Run registers the method parameters and returns an error value declared by a testify mock setup.
// This can be used to assert a correct method call.
func (e *EventTestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

// ListenTarget registers the method parameters, which can be asserted in unit tets.
// It also generates a network event object containing a cached result to test correct event listening behaviour.
func (e *EventTestExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
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
				ConnectStart:      ConnectStart,
				ReceiveHeadersEnd: ReceiveHeadersEnd,
			},
		},
	}

	fn(ev)
}
