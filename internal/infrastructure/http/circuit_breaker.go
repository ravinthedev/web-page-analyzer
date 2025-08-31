package http

import (
	"net/http"
	"time"
	"webpage-analyzer/pkg/logger"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CircuitBreakerHTTPClient struct {
	client *http.Client
	cb     *gobreaker.CircuitBreaker
	logger logger.Logger
}

func NewCircuitBreakerHTTPClient(timeout time.Duration, logger logger.Logger) *CircuitBreakerHTTPClient {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	settings := gobreaker.Settings{
		Name:        "http-client",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Info("Circuit breaker state changed",
				zap.String("name", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	return &CircuitBreakerHTTPClient{
		client: client,
		cb:     cb,
		logger: logger,
	}
}

func (c *CircuitBreakerHTTPClient) Get(url string) (*http.Response, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.Get(url)
	})

	if err != nil {
		c.logger.Error("Circuit breaker request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, err
	}

	return result.(*http.Response), nil
}

func (c *CircuitBreakerHTTPClient) Head(url string) (*http.Response, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.Head(url)
	})

	if err != nil {
		c.logger.Error("Circuit breaker HEAD request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, err
	}

	return result.(*http.Response), nil
}

func (c *CircuitBreakerHTTPClient) Do(req *http.Request) (*http.Response, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.Do(req)
	})

	if err != nil {
		c.logger.Error("Circuit breaker request failed",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return result.(*http.Response), nil
}

func (c *CircuitBreakerHTTPClient) GetState() gobreaker.State {
	return c.cb.State()
}

func (c *CircuitBreakerHTTPClient) GetCounts() gobreaker.Counts {
	return c.cb.Counts()
}
