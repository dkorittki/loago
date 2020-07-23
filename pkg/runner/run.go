package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

func Call(ctx context.Context, url string) (time.Duration, int, string, bool, error) {
	v := FromContext(ctx)

	url = strings.TrimSuffix(url, "/")

	switch v.(type) {
	case *ChromeRunner:
		return runChrome(ctx, url)
	case *FakeRunner:
		return runFake(ctx, url)
	}

	return 0, 0, "", false, InvalidContextError
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
		return 0, 0, "", false, errors.New("did not receive any network events")
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

	if code == 0 {
		return 0, 0, "", false, NoNetworkEventFoundError
	}

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

	time.Sleep(50 * time.Millisecond)

	return 50 * time.Millisecond, 200, "HTTP OK", false, nil
}

// FromContext extracts the runner instance from ctx.
func FromContext(ctx context.Context) Runner {
	v, _ := ctx.Value(contextKey{}).(Runner)
	return v
}
