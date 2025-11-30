package jira

import (
	"github.com/sssidkn/jira-connector/pkg/logger"
	"github.com/sssidkn/jira-connector/pkg/ratelimiter"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	config     Config
	logger     logger.Logger
	rl         *ratelimiter.RateLimiter
	maxDelay   time.Duration
	startDelay time.Duration
}

func NewClient(options ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{},
	}
	for _, opt := range options {
		opt(client)
	}
	client.rl = ratelimiter.NewRateLimiter(client.startDelay, client.maxDelay)
	return client
}
