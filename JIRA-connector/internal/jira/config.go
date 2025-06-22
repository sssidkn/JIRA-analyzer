package jira

import (
	"jira-connector/pkg/logger"
	"net/http"
)

type Config struct {
	BaseURL        string `yaml:"BaseURL" env:"BASE_URL"`
	VersionAPI     string `yaml:"VersionAPI" env:"VERSION_API"`
	MaxConnections int    `yaml:"MaxConnections" env:"RETRY_COUNT"`
	MaxProcesses   int    `yaml:"MaxProcesses" env:"MAX_PROCESSES"`
	RetryCount     int    `yaml:"RetryCount" env:"RETRY_COUNT"`
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
		c.logger = log
	}
}

func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}
