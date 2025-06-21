package jira

import (
	"jira-connector/pkg/logger"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	config     Config
	logger     logger.Logger
}

func NewClient(options ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{},
	}
	for _, opt := range options {
		opt(client)
	}
	return client
}
