package runner

import (
	"context"
	"loago-worker/internal/testing/browser"
	"reflect"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"

	"github.com/chromedp/chromedp"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

const TestingVal = "testing"

type TestingKey struct{}

// isChromeRunnerContext tests if ctx is a valid
// chrome runner context.
// ctx must be created with context.WithValue(ctx, TestingKey{}, TestingVal).
func isChromeRunnerContext(ctx context.Context) bool {
	// Test if ctx is derived from parameter context.
	val := ctx.Value(TestingKey{})
	if val != TestingVal {
		return false
	}

	// Test if ctx is derived from runner context.
	val = FromContext(ctx)
	_, ok := val.(*ChromeRunner)
	if !ok {
		return false
	}

	// Test if ctx is derived from chromedp context.
	return isChromeDPContext(ctx)
}

// isChromeDPContext returns true, if ctx is a chromedp derived context
// or false if not.
func isChromeDPContext(ctx context.Context) bool {
	val := chromedp.FromContext(ctx)
	if val == nil {
		return false
	}

	return true
}

// isNetworkEnableAction checks, if a is a
// network.Enable() chromedp action.
func isNetworkEnableAction(a []chromedp.Action) bool {
	if len(a) != 1 {
		return false
	}

	if reflect.TypeOf(&network.EnableParams{}) != reflect.TypeOf(a[0]) {
		return false
	}

	return true
}

func TestNewChromeRunner(t *testing.T) {
	id := 1
	e := browser.NewEventTestExecutor()

	r := NewChromeRunner(id, e)

	assert.Equal(t, id, r.ID)
	assert.Equal(t, e, r.Executor)
}

func TestChromeRunner_WithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), TestingKey{}, TestingVal)
	e := browser.NewEventTestExecutor()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkEnableAction)).
		Return(nil).
		Once()
	e.On("ListenTarget",
		mock.MatchedBy(isChromeDPContext),
		mock.AnythingOfType("func(interface {})")).
		Once()

	r := NewChromeRunner(1, e)
	returnedCtx := r.WithContext(ctx)

	e.AssertExpectations(t)
	assert.Equal(t, TestingVal, returnedCtx.Value(TestingKey{}))

	assert.Len(t, r.networkEventChan, 1)
	ev := <-r.networkEventChan
	assert.Equal(t, network.ResourceTypeDocument, ev.Type)
	assert.Equal(t, network.RequestID("testing"), ev.RequestID)

	_, ok := FromContext(returnedCtx).(*ChromeRunner)
	assert.True(t, ok)
}

func TestCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), TestingKey{}, TestingVal))
	e := browser.NewEventTestExecutor()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkEnableAction)).
		Return(nil).
		Once()
	e.On("ListenTarget",
		mock.MatchedBy(isChromeDPContext),
		mock.AnythingOfType("func(interface {})"))

	r := NewChromeRunner(1, e)
	_ = r.WithContext(ctx)

	// Todo: Find a better way than sleep() to wait for cancel to finish.
	cancel()
	time.Sleep(300 * time.Millisecond)

	// Test closed network buffer channel.
	assert.Panics(t, func() {
		r.networkEventChan <- &network.EventResponseReceived{}
	})
}
