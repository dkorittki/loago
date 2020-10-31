package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

// Call executes an request on url using the runner context.
// ctx must be a valid runner context created with WithContext method of a runner instance.
// It returns the response time, HTTP response code, the HTTP response message an
// a boolean indicating if the content comes from a browser cache.
//
// If an error occurred while performing the request an error is returned with zero values
// on all other return values.
func Call(ctx context.Context, url string) (time.Duration, int, string, bool, error) {
	v := FromContext(ctx)

	url = strings.TrimSuffix(url, "/")

	switch v.(type) {
	case *ChromeRunner:
		return runChrome(ctx, url)
	case *FakeRunner:
		return runFake(ctx, url)
	}

	return 0, 0, "", false, ErrInvalidContext
}

func runChrome(ctx context.Context, url string) (time.Duration, int, string, bool, error) {
	r := FromContext(ctx).(*ChromeRunner)

	log.Debug().
		Str("component", "runner").
		Int("id", r.ID).
		Str("type", fmt.Sprintf("%T", r)).
		Str("url", url).
		Msg("call url")

	err := r.Executor.Run(ctx, network.Enable())
	if err != nil {
		return 0, 0, "", false, err
	}

	err = r.Executor.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Stop(),
	)

	if err != nil {
		return 0, 0, "", false, err
	}

	err = r.Executor.Run(ctx, network.Disable())
	if err != nil {
		return 0, 0, "", false, err
	}

	var (
		ttfb   time.Duration
		code   int
		msg    string
		cached bool
	)

	// Read received network events from runner buffer,
	// read network stats and parse ttfb.
	if len(r.networkEventChan) == 0 {
		return 0, 0, "", false, ErrNoNetworkEventFound
	}

	func() {
		for {
			select {
			case ev := <-r.networkEventChan:
				if strings.TrimSuffix(ev.Response.URL, "/") == url {
					log.Debug().
						Str("component", "runner").
						Int("id", r.ID).
						Interface("ev", ev.Response.Timing).
						Msg("received base url network event")

					code = int(ev.Response.Status)
					msg = ev.Response.StatusText

					if ev.Response.Timing.ConnectStart == -1 {
						ttfb = 0
						cached = true
					} else {
						ttfb = time.Duration(ev.Response.Timing.ReceiveHeadersEnd-
							ev.Response.Timing.ConnectStart) * time.Millisecond
					}
				}
			default:
				return
			}
		}
	}()

	return ttfb, code, msg, cached, nil
}

func runFake(ctx context.Context, url string) (time.Duration, int, string, bool, error) {
	r := FromContext(ctx).(*FakeRunner)

	log.Debug().
		Str("component", "runner").
		Uint("id", uint(*r)).
		Str("type", fmt.Sprintf("%T", r)).
		Str("url", url).
		Msg("call url")

	select {
	case <-time.After(50 * time.Millisecond):
		return 50 * time.Millisecond, 200, "OK", false, nil
	case <-ctx.Done():
		return 0, 0, "", false, context.Canceled
	}
}
