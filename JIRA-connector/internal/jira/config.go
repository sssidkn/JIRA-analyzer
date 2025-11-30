package jira

import (
	"github.com/sssidkn/jira-connector/pkg/logger"
	"net/http"
	"time"
)

type Config struct {
	BaseURL        string `yaml:"BaseURL" env:"BASE_URL"`
	VersionAPI     string `yaml:"VersionAPI" env:"VERSION_API"`
	MaxConnections int    `yaml:"MaxConnections" env:"RETRY_COUNT"`
	MaxProcesses   int    `yaml:"MaxProcesses" env:"MAX_PROCESSES"`
	MaxDelay       int    `yaml:"MaxDelay" env:"MAX_DELAY"`
	StartDelay     int    `yaml:"StartDelay" env:"START_DELAY"`
	MaxResults     int    `yaml:"MaxResults" env:"MAX_RESULTS"`
}

type Option func(*Client)

func WithConfig(cfg Config) func(*Client) {
	return func(c *Client) {
		c.httpClient.Transport = &http.Transport{
			MaxConnsPerHost:     cfg.MaxConnections,
			MaxIdleConnsPerHost: cfg.MaxConnections,
		}
		c.config = cfg
	}
}

func WithLogger(log logger.Logger) func(*Client) {
	if log == nil {
		log = logger.NewLogrusLogger()
		log.SetLevel(logger.LevelInfo)
	}
	return func(c *Client) {
		c.logger = log.With(logger.Field{Key: "module", Value: "JIRA_API_Client"})
	}
}

func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

func WithStartDelay(delay int) func(client *Client) {
	return func(c *Client) {
		c.startDelay = time.Duration(delay) * time.Second
	}
}

func WithMaxDelay(delay int) func(client *Client) {
	return func(c *Client) {
		c.maxDelay = time.Duration(delay) * time.Second
	}
}
