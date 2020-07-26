package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

type Executor interface {
	Run(ctx context.Context, actions ...chromedp.Action) error
	ListenTarget(ctx context.Context, fn func(ev interface{}))
}

type ChromeDPExecutor struct{}

func New() *ChromeDPExecutor {
	return &ChromeDPExecutor{}
}

func (e *ChromeDPExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	return chromedp.Run(ctx, actions...)
}

func (e *ChromeDPExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
	chromedp.ListenTarget(ctx, fn)
}
