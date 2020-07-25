package loadtest

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	chromedpexecutor "github.com/dkorittki/loago-worker/internal/executor/browser"
	"github.com/dkorittki/loago-worker/pkg/runner"

	"github.com/rs/zerolog/log"
)

var (
	InvalidRunnerTypeError = errors.New("invalid runner type")
	InvalidWaitBoundaries  = errors.New("max wait duration is bigger than min wait duration")
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Run(ctx context.Context,
	browserType BrowserType,
	endpoints []*Endpoint,
	minWait, maxWait time.Duration,
	amount int,
	results chan EndpointResult) error {
	log.Info().Str("component", "loadtest_service").Msg("starting a new loadtest")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// create temporary slice random selection of endpoints.
	var e []*Endpoint
	for i, v := range endpoints {
		for j := 0; j < int(v.Weight); j++ {
			e = append(e, endpoints[i])
		}
	}

	errChan := make(chan error, amount)
	wgDone := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(amount)

	for i := 0; i < amount; i++ {
		var r runner.Runner
		switch browserType {
		case BrowserTypeFake:
			r = runner.NewFakeRunner(i)
		case BrowserTypeChrome:
			e := chromedpexecutor.New()
			r = runner.NewChromeRunner(i, e)
		default:
			return InvalidRunnerTypeError
		}

		runnerCtx := r.WithContext(ctx)

		// Start a new schedule for this specific runner.
		id := i
		go func() {
			err := schedule(runnerCtx, id, e, minWait, maxWait, results, &wg)
			if err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		log.Info().Msg("schedules finished work successfully")
		return nil
	case err := <-errChan:
		return err
	}

}

func schedule(ctx context.Context, id int, endpoints []*Endpoint, minWait, maxWait time.Duration, results chan EndpointResult, wg *sync.WaitGroup) error {
	log.Info().
		Str("component", "schedule").
		Int("id", id).
		Msg("start new schedule")

	defer wg.Done()

	for {
		err := sleepBetween(minWait, maxWait)
		if err != nil {
			return err
		}

		select {
		default:
			url := endpoints[rand.Intn(len(endpoints))].URL
			ttfb, code, msg, cached, err := runner.Call(ctx, url)

			if err != nil {
				if err == context.Canceled {
					log.Debug().
						Str("component", "schedule").
						Int("id", id).
						Msg("context canceld mid request")

					return nil
				} else if err == context.DeadlineExceeded {
					log.Warn().
						Str("component", "schedule").
						Int("id", id).
						Msg("request timed out")
					continue
				}

				return err
			}

			results <- EndpointResult{
				URL:               url,
				HTTPStatusCode:    code,
				HTTPStatusMessage: msg,
				TTFB:              ttfb,
				Cached:            cached,
			}
		case <-ctx.Done():
			log.Info().
				Str("component", "schedule").
				Int("id", id).
				Msg("stop schedule")

			return nil
		}
	}
}

func sleepBetween(min, max time.Duration) error {
	var z time.Duration

	if min == max {
		z = min
	} else if min > max {
		return InvalidWaitBoundaries
	} else {
		z = time.Duration(int64(min) + rand.Int63n(int64(max-min)))
	}

	time.Sleep(z)
	return nil
}
