package runner

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dkorittki/loago-worker/internal/pkg/testing/browser"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func isNavigateAction(a []chromedp.Action) bool {
	if len(a) != 2 {
		return false
	}

	if reflect.TypeOf(chromedp.Navigate("")) != reflect.TypeOf(a[0]) {
		return false
	}

	if reflect.TypeOf(&page.StopLoadingParams{}) != reflect.TypeOf(a[1]) {
		return false
	}

	return true
}

func isNetworkDisableAction(a []chromedp.Action) bool {
	if len(a) != 1 {
		return false
	}

	if reflect.TypeOf(&network.DisableParams{}) != reflect.TypeOf(a[0]) {
		return false
	}

	return true
}

func TestCall_FakeRunner(t *testing.T) {
	r := NewFakeRunner(1)
	ctx := r.WithContext(context.Background())

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	assert.NoError(t, err)
	assert.Equal(t, 50*time.Millisecond, ttfb)
	assert.Equal(t, 200, code)
	assert.Equal(t, "OK", msg)
	assert.False(t, cached)
}

func TestCall_ChromeRunner(t *testing.T) {
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
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNavigateAction)).
		Return(nil).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkDisableAction)).
		Return(nil).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(browser.ReceiveHeadersEnd-browser.ConnectStart)*time.Millisecond, ttfb)
	assert.Equal(t, int(browser.Status), code)
	assert.Equal(t, browser.StatusText, msg)
	assert.False(t, cached)
}

func TestCall_ChromeRunner_ErrorOnNetworkEnable(t *testing.T) {
	e := browser.NewEventTestExecutor()
	e.On("ListenTarget",
		mock.MatchedBy(isChromeDPContext),
		mock.AnythingOfType("func(interface {})")).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkEnableAction)).
		Return(errors.New("test network enable error")).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	if assert.Error(t, err) {
		assert.Equal(t, "test network enable error", err.Error())
	}
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, 0, code)
	assert.Equal(t, "", msg)
	assert.False(t, cached)
}

func TestCall_ChromeRunner_ErrorOnNavigateAction(t *testing.T) {
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
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNavigateAction)).
		Return(errors.New("test navigate error")).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	if assert.Error(t, err) {
		assert.Equal(t, "test navigate error", err.Error())
	}
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, 0, code)
	assert.Equal(t, "", msg)
	assert.False(t, cached)
}

func TestCall_ChromeRunner_ErrorOnNetworkDisable(t *testing.T) {
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
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNavigateAction)).
		Return(nil).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkDisableAction)).
		Return(errors.New("test network disable error")).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	if assert.Error(t, err) {
		assert.Equal(t, "test network disable error", err.Error())
	}
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, 0, code)
	assert.Equal(t, "", msg)
	assert.False(t, cached)
}

func TestCall_ChromeRunner_EmptyNetworkEventBuffer(t *testing.T) {
	e := browser.NewTestExecutor()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkEnableAction)).
		Return(nil).
		Once()
	e.On("ListenTarget",
		mock.MatchedBy(isChromeDPContext),
		mock.AnythingOfType("func(interface {})")).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNavigateAction)).
		Return(nil).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkDisableAction)).
		Return(nil).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	if assert.Error(t, err) {
		assert.Equal(t, NoNetworkEventFoundError, err)
	}
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, 0, code)
	assert.Equal(t, "", msg)
	assert.False(t, cached)

}

func TestCall_ChromeRunner_CachedHTTPResponse(t *testing.T) {
	e := browser.NewCachedTestExecutor()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkEnableAction)).
		Return(nil).
		Once()
	e.On("ListenTarget",
		mock.MatchedBy(isChromeDPContext),
		mock.AnythingOfType("func(interface {})")).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNavigateAction)).
		Return(nil).
		Once()
	e.On("Run",
		mock.MatchedBy(isChromeRunnerContext),
		mock.MatchedBy(isNetworkDisableAction)).
		Return(nil).
		Once()

	r := NewChromeRunner(1, e)
	ctx := r.WithContext(context.WithValue(context.Background(), TestingKey{}, TestingVal))

	ttfb, code, msg, cached, err := Call(ctx, "http://foo.bar")

	e.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, int(browser.Status), code)
	assert.Equal(t, browser.StatusText, msg)
	assert.True(t, cached)
}

func TestCall_InvalidRunner(t *testing.T) {
	ttfb, code, msg, cached, err := Call(context.Background(), "http://foo.bar")

	assert.Error(t, err)
	assert.Equal(t, time.Duration(0), ttfb)
	assert.Equal(t, 0, code)
	assert.Equal(t, "", msg)
	assert.False(t, cached)
}
