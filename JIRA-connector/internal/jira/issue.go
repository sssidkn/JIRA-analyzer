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
	"time"

	"github.com/google/uuid"
)

func (c *Client) getTotalIssuesCount(ctx context.Context, projectKey string) (int, error) {
	jql := fmt.Sprintf("project=%s", projectKey)
	c.logger.Info("starting getting total issues count")
	params := url.Values{
		"jql":        []string{jql},
		"startAt":    []string{"0"},
		"maxResults": []string{"0"},
	}

	var result struct {
		Total int `json:"total"`
	}

	endpoint := c.buildURL("/search", params)
	if err := c.doRequest(ctx, endpoint, &result); err != nil {
		return 0, fmt.Errorf("failed to get total issues count: %w", err)
	}
	c.logger.Info(fmt.Sprintf("finished getting total issues count. Count: %d", result.Total))
	return result.Total, nil
}

func (c *Client) getIssuesCountAfter(ctx context.Context, projectKey string, lastUpdate time.Time) (int, error) {
	jql := fmt.Sprintf("project=%s AND updated > \"%s\"", projectKey,
		lastUpdate.UTC().Format("2006/01/02"))

	params := url.Values{
		"jql":        []string{jql},
		"maxResults": []string{"0"},
	}

	var result struct {
		Total int `json:"total"`
	}

	endpoint := c.buildURL("/search", params)
	err := c.doRequest(ctx, endpoint, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to get issues count: %w", err)
	}
	return result.Total, nil
}

func (c *Client) getIssuesBy(ctx context.Context, total int, params url.Values) (*[]models.JiraIssue, error) {
	pageSize := c.config.MaxResults

	totalPages := (total + pageSize - 1) / pageSize
	c.logger.Debug(fmt.Sprintf("Total pages: %d", totalPages))

	tasks := make(chan int, totalPages)
	results := make(chan *result, totalPages)
	var wg sync.WaitGroup
	link := c.buildURL("/search", params)

	for i := 0; i < c.config.MaxConnections; i++ {
		wg.Add(1)

		ctx = context.WithValue(ctx, "worker_id", uuid.New())
		go c.issuePageWorker(ctx, tasks, results, &wg, link)
	}

	for page := 0; page < totalPages; page++ {
		tasks <- page
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	allIssues := make([]models.JiraIssue, 0, total)
	var errors []error

	for res := range results {
		if res.err != nil {
			errors = append(errors, fmt.Errorf("page %d: %w", res.page, res.err))
			continue
		}
		allIssues = append(allIssues, res.issues...)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("%d errors occurred, first error: %w", len(errors), errors[0])
	}
	return &allIssues, nil
}

func (c *Client) getAllIssues(ctx context.Context, projectKey string, total int) (*[]models.JiraIssue, error) {
	pageSize := c.config.MaxResults

	allIssues, err := c.getIssuesBy(ctx, total, url.Values{
		"jql":        []string{fmt.Sprintf("project=%s", projectKey)},
		"maxResults": []string{fmt.Sprintf("%d", pageSize)},
		"expand":     []string{"changelog"},
		"fields": []string{`summary,description,issuetype,priority,
			status,creator,assignee,created,updated,resolutiondate,worklog,timetracking`},
	})
	if err != nil {
		return nil, err
	}
	return allIssues, nil
}

func (c *Client) getIssuesPage(ctx context.Context, link string) ([]models.JiraIssue, error) {
	var result struct {
		Issues []models.JiraIssue `json:"issues"`
	}

	err := c.doRequest(ctx, link, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues page: %w", err)
	}

	return result.Issues, nil
}

func (c *Client) issuePageWorker(ctx context.Context,
	tasks <-chan int,
	results chan<- *result,
	wg *sync.WaitGroup, link string) {
	defer wg.Done()
	pageSize := c.config.MaxResults
	for page := range tasks {
		startAt := page * pageSize
		c.logger.Debug("Processing page",
			logger.Field{Key: "page_size", Value: pageSize},
			logger.Field{Key: "worker", Value: ctx.Value("worker_id")},
			logger.Field{Key: "page", Value: page})
		issues, err := c.getIssuesPage(ctx, link+fmt.Sprintf("&startAt=%d", startAt))

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
