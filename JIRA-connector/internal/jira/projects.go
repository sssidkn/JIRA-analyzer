package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jira-connector/internal/models"
	"jira-connector/pkg/logger"
	"net/http"
	"net/url"
	"time"
)

// GetProject returns project information
func (c *Client) GetProject(ctx context.Context, projectKey string) (*models.JiraProject, error) {
	endpoint := fmt.Sprintf("%s%s/project/%s?expand=insight,description,lead",
		c.config.BaseURL, c.config.VersionAPI, projectKey)
	log := c.logger.With(
		logger.Field{Key: "project_key", Value: projectKey},
		logger.Field{Key: "project_url", Value: endpoint},
	)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var project models.JiraProject

	if err = json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	project.Self = c.config.BaseURL + "/projects/" + project.Key
	total, err := c.getTotalIssuesCount(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get total issues count: %w", err)
	}
	project.TotalIssueCount = total
	if project.TotalIssueCount == 0 {
		return &project, nil
	}

	var issues *[]models.JiraIssue

	log.Info("Fetched issues count", logger.Field{Key: "total", Value: project.TotalIssueCount})
	log.Info("Fetching issues")
	issues, err = c.getAllIssues(ctx, projectKey, project.TotalIssueCount)
	log.Info("Fetching issues ended")
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	project.Issues = *issues
	return &project, nil
}

// UpdateProject updates project
func (c *Client) UpdateProject(ctx context.Context, projectKey string, lastUpdate time.Time) (*[]models.JiraIssue, error) {
	total, err := c.getIssuesCountAfter(ctx, projectKey, lastUpdate)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("project=%s AND updated > \"%s\"", projectKey,
		lastUpdate.UTC().Format("2006/01/02"))

	params := url.Values{
		"jql":        []string{query},
		"maxResults": []string{fmt.Sprintf("%d", c.config.MaxResults)},
		"expand":     []string{"changelog"},
		"fields": []string{`summary,description,issuetype,priority,
			status,creator,assignee,created,updated,resolutiondate,worklog,timetracking`},
	}

	issues, err := c.getIssuesBy(ctx, total, params)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// GetProjects returns projects
func (c *Client) GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error) {
	c.logger.Info("starting getting projects")
	response, err := http.Get(c.config.BaseURL + c.config.VersionAPI + "/project")
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jiraProjects []models.ProjectInfo
	err = json.Unmarshal(body, &jiraProjects)
	for i := 0; i < len(jiraProjects); i++ {
		jiraProjects[i].Self = c.config.BaseURL + "/projects/" + jiraProjects[i].Key
	}
	if err != nil {
		return nil, err
	}

	return jiraProjects, nil
}
