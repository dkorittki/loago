package browser

import (
	"context"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

type CachedTestExecutor struct {
	mock.Mock
}

func NewCachedTestExecutor() *CachedTestExecutor {
	return &CachedTestExecutor{}
}

func (e *CachedTestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

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

	return
}
