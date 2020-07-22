package runner

import (
	"context"
)

type FakeRunner uint

func NewFakeRunner(id uint) *FakeRunner {
	f := FakeRunner(id)
	return &f
}

func (r *FakeRunner) WithContext(ctx context.Context) context.Context {
	runnerCtx := context.WithValue(ctx, contextKey{}, r)
	return runnerCtx
}
