package jira

import (
	"context"
	"fmt"
	"jira-connector/internal/models"
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

	for i := 0; i < c.config.MaxConnections; i++ {
		wg.Add(1)

		ctx = context.WithValue(ctx, "worker_id", uuid.New())
		go c.issuePageWorker(ctx, tasks, results, &wg, params)
	}

	for page := 0; page < totalPages; page++ {
		tasks <- page
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allIssues []models.JiraIssue
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

func (c *Client) getIssuesPage(ctx context.Context, params url.Values) ([]models.JiraIssue, error) {
	var result struct {
		Issues []models.JiraIssue `json:"issues"`
	}

	endpoint := c.buildURL("/search", params)
	err := c.doRequest(ctx, endpoint, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues page (startAt: %d): %w", params.Get("startAt"), err)
	}

	return result.Issues, nil
}
