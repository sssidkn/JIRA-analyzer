package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"
	"net/url"
	"sync"
)

func (c *Client) issuePageWorker(ctx context.Context,
	tasks <-chan int,
	results chan<- *result,
	wg *sync.WaitGroup, params url.Values) {
	defer wg.Done()
	pageSize := c.config.MaxResults
	for page := range tasks {
		startAt := page * pageSize
		c.logger.Debug("Processing page",
			logger.Field{Key: "project_key", Value: params.Get("projectKey")},
			logger.Field{Key: "page_size", Value: pageSize},
			logger.Field{Key: "worker", Value: ctx.Value("worker_id")},
			logger.Field{Key: "page", Value: page})
		params.Set("startAt", fmt.Sprintf("%d", startAt))
		issues, err := c.getIssuesPage(ctx, params)

		select {
		case results <- &result{issues: issues, err: err, page: page}:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) buildURL(endpoint string, params url.Values) string {
	return fmt.Sprintf("%s%s%s?%s", c.config.BaseURL, c.config.VersionAPI, endpoint, params.Encode())
}

func (c *Client) doRequest(ctx context.Context, url string, result interface{}) error {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err = json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

type result struct {
	issues []models.JiraIssue `json:"issues"`
	err    error
	page   int
}
