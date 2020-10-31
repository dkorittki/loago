package runner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFakeRunner(t *testing.T) {
	f := NewFakeRunner(1)
	f2 := FakeRunner(1)

	assert.Equal(t, &f2, f)
}

func TestFakeRunner_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	f := NewFakeRunner(1)
	f2 := FakeRunner(1)
	ctx2 := f.WithContext(ctx)
	r, ok := ctx2.Value(contextKey{}).(*FakeRunner)
	require.True(t, ok)

	assert.Equal(t, &f2, r)

	cancel()
	<-ctx2.Done()
}
