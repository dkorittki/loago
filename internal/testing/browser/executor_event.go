package browser

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

var (
	URL               = "http://foo.bar"
	Status            = int64(200)
	StatusText        = "Testing OK"
	ConnectStart      = float64(60)
	ReceiveHeadersEnd = float64(100)
)

type EventTestExecutor struct {
	mock.Mock
}

func NewEventTestExecutor() *EventTestExecutor {
	return &EventTestExecutor{}
}

func (e *EventTestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

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

	return
}
