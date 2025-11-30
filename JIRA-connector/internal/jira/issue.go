package jira

import (
	"context"
	"errors"
	"fmt"
	"github.com/sssidkn/jira-connector/internal/models"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
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
	threadsCount := c.config.MaxProcesses

	totalPages := (total + pageSize - 1) / pageSize
	c.logger.Debug(fmt.Sprintf("Total pages: %d", totalPages))

	pages := make(chan int, threadsCount)
	results := make(chan []models.JiraIssue, threadsCount)

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.SetLimit(threadsCount + 1)

	link := c.buildURL("/search", params)

	errGroup.Go(func() error {
		defer close(pages)
		for page := 0; page < totalPages; page++ {
			select {
			case pages <- page:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	for i := 0; i < threadsCount; i++ {
		errGroup.Go(func() error {
			return c.issuePageWorker(ctx, pages, results, link)
		})
	}

	go func() {
		errGroup.Wait()
		close(results)
	}()

	allIssues := make([]models.JiraIssue, 0, total)
	for res := range results {
		allIssues = append(allIssues, res...)
	}

	if err := errGroup.Wait(); err != nil {
		return nil, err
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
		return nil, err
	}

	return result.Issues, nil
}

func (c *Client) issuePageWorker(ctx context.Context, pages chan int,
	issuePages chan<- []models.JiraIssue, link string) error {

	pageSize := c.config.MaxResults
	for {
		if err := c.waitIfPaused(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case page, ok := <-pages:
			if !ok {
				return nil
			}

			startAt := page * pageSize
			c.logger.Debug("Processing page",
				logger.Field{Key: "page_size", Value: pageSize},
				logger.Field{Key: "page", Value: page})

			issues, err := c.getIssuesPage(ctx, link+fmt.Sprintf("&startAt=%d", startAt))

			var apiErr *APIError
			if errors.As(err, &apiErr) {
				if apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500 {
					c.logger.Info("API rate limit exceeded", logger.Field{Key: "Error", Value: apiErr.Error()})
					c.rl.Pause()
					go func() { pages <- page }()
					continue
				} else {
					return err
				}
			}

			c.rl.Reset()

			select {
			case issuePages <- issues:
				c.logger.Debug("Processed page",
					logger.Field{Key: "page_size", Value: pageSize},
					logger.Field{Key: "page", Value: page})
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (c *Client) waitIfPaused(ctx context.Context) error {
	for {
		paused, duration := c.rl.ShouldPause()
		if !paused {
			return nil
		}
		if duration >= c.maxDelay {
			return errors.New("exceeded max delay")
		}
		c.logger.Info("Request paused due to rate limiting",
			logger.Field{Key: "retry_after", Value: duration.String()})

		select {
		case <-time.After(duration):
			continue
		case <-ctx.Done():
			return ctx.Err()
		case <-c.rl.NotifyPause():
			continue
		}
	}
}
