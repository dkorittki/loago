package browser

import (
	"context"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/mock"
)

type TestExecutor struct {
	mock.Mock
}

func NewTestExecutor() *TestExecutor {
	return &TestExecutor{}
}

func (e *TestExecutor) Run(ctx context.Context, actions ...chromedp.Action) error {
	args := e.Called(ctx, actions)
	return args.Error(0)
}

func (e *TestExecutor) ListenTarget(ctx context.Context, fn func(ev interface{})) {
	e.Called(ctx, fn)
	return
}
