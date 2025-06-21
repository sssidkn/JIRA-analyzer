package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jira-connector/internal/models"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Config struct {
	BaseURL        string `yaml:"BaseURL" env:"BASE_URL"`
	MaxConnections int    `yaml:"MaxConnections" env:"RETRY_COUNT"`
	RetryCount     int    `yaml:"RetryCount" env:"RETRY_COUNT"`
	MaxResults     int    `yaml:"MaxResults" env:"MAX_RESULTS"`
}

type Client struct {
	httpClient *http.Client
	config     Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    cfg.MaxConnections,
				MaxConnsPerHost: cfg.MaxConnections,
			},
		},
		config: cfg,
	}
}

func (c *Client) GetProject(ctx context.Context, projectKey string) (*models.JiraProject, error) {
	url := fmt.Sprintf("%s/project/%s", c.config.BaseURL, projectKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var project models.JiraProject
	if err = json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	var issues []models.JiraIssue

	total, err := c.getTotalIssuesCount(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get total issues count: %w", err)
	}
	fmt.Println(total)
	issues, err = c.getAllIssues(ctx, projectKey, total)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	project.Issues = issues
	return &project, nil
}

func (c *Client) getTotalIssuesCount(ctx context.Context, projectKey string) (int, error) {
	jql := fmt.Sprintf("project=%s", projectKey)
	params := url.Values{
		"jql":        []string{jql},
		"startAt":    []string{"0"},
		"maxResults": []string{"1"},
	}

	var result struct {
		Total int `json:"total"`
	}

	endpoint := c.buildURL("/search", params)
	if err := c.doRequest(ctx, endpoint, &result); err != nil {
		return 0, fmt.Errorf("failed to get total issues count: %w", err)
	}

	return result.Total, nil
}

func (c *Client) GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error) {
	response, err := http.Get(c.config.BaseURL + "/project")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var jiraProjects []models.ProjectInfo
	err = json.Unmarshal(body, &jiraProjects)

	if err != nil {
		return nil, err
	}

	var projects []models.ProjectInfo

	projectsCount := 0

	for _, elem := range jiraProjects {
		if strings.Contains(strings.ToLower(elem.Name), strings.ToLower(search)) {
			projectsCount++
			projects = append(projects, models.ProjectInfo{
				ID:   elem.ID,
				Name: elem.Name,
				Key:  elem.Key,
			})
		}
	}

	startIndex := limit * (page - 1)
	endIndex := startIndex + limit
	if endIndex >= len(projects) {
		endIndex = len(projects)
	}

	return projects, nil
}

func (c *Client) issuePageWorker(ctx context.Context, projectKey string, pageSize int,
	wg *sync.WaitGroup, tasks <-chan int, results chan<- *result) {
	defer wg.Done()

	for page := range tasks {
		startAt := page * pageSize
		issues, err := c.getIssuesPage(ctx, projectKey, startAt, pageSize)

		select {
		case results <- &result{issues: issues, err: err, page: page}:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) getAllIssues(ctx context.Context, projectKey string, total int) ([]models.JiraIssue, error) {
	pageSize := c.config.MaxResults

	totalPages := (total + pageSize - 1) / pageSize

	tasks := make(chan int, totalPages)
	results := make(chan *result, totalPages)
	var wg sync.WaitGroup

	for i := 0; i < c.config.MaxConnections; i++ {
		wg.Add(1)
		go c.issuePageWorker(ctx, projectKey, pageSize, &wg, tasks, results)
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

	return allIssues, nil
}

func (c *Client) getIssuesPage(ctx context.Context, projectKey string, startAt, maxResults int) ([]models.JiraIssue, error) {
	jql := fmt.Sprintf("project=%s", projectKey)
	params := url.Values{
		"jql":        []string{jql},
		"startAt":    []string{fmt.Sprintf("%d", startAt)},
		"maxResults": []string{fmt.Sprintf("%d", maxResults)},
		"expand":     []string{"changelog"},
	}

	var result struct {
		Issues []models.JiraIssue `json:"issues"`
	}

	endpoint := c.buildURL("/search", params)
	err := c.doRequest(ctx, endpoint, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get issues page (startAt: %d): %w", startAt, err)
	}

	return result.Issues, nil
}

func (c *Client) buildURL(endpoint string, params url.Values) string {
	return fmt.Sprintf("%s%s?%s", c.config.BaseURL, endpoint, params.Encode())
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
