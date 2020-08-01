package loadtest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, &Service{}, New())
}

func TestService_Run(t *testing.T) {
	type input struct {
		minWait     time.Duration
		maxWait     time.Duration
		endpoints   []*Endpoint
		amount      int
		browserType BrowserType
	}

	type output struct {
		greaterOrEqual int
		lessOrEqual    int
		err            error
	}

	vars := []struct {
		name string
		in   input
		out  output
	}{
		{
			name: "Single",
			in: input{
				minWait: time.Second,
				maxWait: time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      1,
				browserType: BrowserTypeFake,
			},
			out: output{
				greaterOrEqual: 9,
				lessOrEqual:    10,
			},
		},
		{
			name: "Double",
			in: input{
				minWait: time.Second,
				maxWait: time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      2,
				browserType: BrowserTypeFake,
			},
			out: output{
				greaterOrEqual: 18,
				lessOrEqual:    20,
			},
		},
		{
			name: "Ten",
			in: input{
				minWait: time.Second,
				maxWait: time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      10,
				browserType: BrowserTypeFake,
			},
			out: output{
				greaterOrEqual: 90,
				lessOrEqual:    100,
			},
		},
		{
			name: "Irregular",
			in: input{
				minWait: time.Second,
				maxWait: 2 * time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      1,
				browserType: BrowserTypeFake,
			},
			out: output{
				greaterOrEqual: 5,
				lessOrEqual:    10,
			},
		},
		{
			name: "InvalidRunner",
			in: input{
				minWait: time.Second,
				maxWait: 2 * time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      1,
				browserType: 2,
			},
			out: output{
				greaterOrEqual: 0,
				lessOrEqual:    0,
				err:            ErrInvalidRunnerType,
			},
		},
		{
			name: "ErrInvalidWaitBoundaries",
			in: input{
				minWait: 2 * time.Second,
				maxWait: time.Second,
				endpoints: []*Endpoint{
					{
						URL:    "http://localhost:8080/url1",
						Weight: 1,
					},
				},
				amount:      1,
				browserType: BrowserTypeFake,
			},
			out: output{
				greaterOrEqual: 0,
				lessOrEqual:    0,
				err:            ErrInvalidWaitBoundaries,
			},
		},
	}

	for _, v := range vars {
		t.Run(v.name, func(t *testing.T) {
			results := make(chan EndpointResult, 1000)
			errChan := make(chan error)
			ctx, cancel := context.WithCancel(context.Background())
			s := New()

			go func() {
				errChan <- s.Run(ctx, v.in.browserType, v.in.endpoints, v.in.minWait, v.in.maxWait, v.in.amount, results)
			}()

			go func() {
				time.Sleep(10 * time.Second)
				cancel()
			}()

			// block until Run() has finished
			err := <-errChan
			close(results)

			assert.Equal(t, v.out.err, err)
			assert.GreaterOrEqual(t, len(results), v.out.greaterOrEqual)
			assert.LessOrEqual(t, len(results), v.out.lessOrEqual)

			for r := range results {
				assert.Equal(t, 50*time.Millisecond, r.TTFB)
				assert.Equal(t, 200, r.HTTPStatusCode)
				assert.Equal(t, "OK", r.HTTPStatusMessage)
				assert.False(t, r.Cached)
			}
		})
	}
}
