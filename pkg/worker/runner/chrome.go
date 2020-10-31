package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/dkorittki/loago/internal/pkg/worker/executor/browser"
	"github.com/rs/zerolog/log"
)

const (
	// CacheDirName is the name of the global directory used for runner caches.
	CacheDirName = "loago_runner"

	// Size of network event channel buffer.
	networkEventChanSize = 300
)

// A ChromeRunner implements the runner interface.
// It interacts with a chrome browser via chromedp.
// See: https://github.com/chromedp/chromedp
type ChromeRunner struct {
	// ID of this runner.
	ID int

	// Browser cache directory path.
	CacheDir string

	// Executor interface for interacting with a browser communication library.
	Executor browser.Executor

	// Buffer for storing network events received from devtools protocols.
	networkEventChan chan *network.EventResponseReceived
}

// NewChromeRunner creates a new chrome runner instance.
func NewChromeRunner(id int, e browser.Executor) *ChromeRunner {
	r := &ChromeRunner{
		ID:               id,
		Executor:         e,
		networkEventChan: make(chan *network.EventResponseReceived, networkEventChanSize),
	}

	return r
}

// WithContext derives a new context from ctx associated with both a runner and
// chromedp configuration. This context can be used as a context to call the Run() method.
// It also creates a new goroutine in background waiting for the context to be closed to
// clean up ressources such as the cache dir.
func (r *ChromeRunner) WithContext(ctx context.Context) context.Context {
	cachedir := filepath.Join(os.TempDir(), CacheDirName, fmt.Sprintf("%d", r.ID))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(cachedir),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	chromedpCtx, _ := chromedp.NewContext(allocCtx)

	r.CacheDir = cachedir
	runnerCtx := context.WithValue(chromedpCtx, contextKey{}, r)

	// Watch context and clean up browser cache once it's canceled
	f := cancel(runnerCtx)
	go f()

	// Create a network event listener and send them into the runner buffer.
	// The Call() method will read and parse from it.
	r.Executor.ListenTarget(chromedpCtx, func(ev interface{}) {
		if netEv, ok := ev.(*network.EventResponseReceived); ok {
			if netEv.Type == network.ResourceTypeDocument {
				r.networkEventChan <- netEv
			}
		}
	})

	return runnerCtx
}

func cancel(ctx context.Context) func() {
	return func() {
		v := FromContext(ctx)
		r := v.(*ChromeRunner)

		<-ctx.Done()
		// close network event buffer.
		close(r.networkEventChan)

		log.Debug().
			Str("component", "runner").
			Int("id", r.ID).
			Str("cachedir", r.CacheDir).
			Msg("delete cache")

		var err error
		for i := 0; i < 10; i++ {
			err = os.RemoveAll(r.CacheDir)
			if err == nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}

		log.Warn().
			Str("component", "runner").
			Int("id", r.ID).
			Err(err).
			Msg("can't delete cache")
	}
}
