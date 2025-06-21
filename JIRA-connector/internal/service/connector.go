package connector

import (
	"context"
	"fmt"
	"jira-connector/internal/models"
)

type Issue = models.JiraIssue
type Project = models.JiraProject
type Option = func(*JiraConnector) error

type Repository interface {
	SaveProject(ctx context.Context, project Project) error
	ProjectExists(ctx context.Context, projectKey string) (bool, error)
	Close() error
}

type APIClient interface {
	GetProject(ctx context.Context, projectKey string) (*Project, error)

	GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error)
}

type JiraConnector struct {
	repo      Repository
	apiClient APIClient
}

func NewJiraConnector(opts ...Option) (*JiraConnector, error) {
	jc := &JiraConnector{}
	var err error
	for _, opt := range opts {
		err = opt(jc)
		if err != nil {
			return nil, err
		}
	}
	return jc, nil
}

func WithRepository(repo Repository) Option {
	return func(jc *JiraConnector) error {
		if repo == nil {
			return fmt.Errorf("ERROR: repo is nil")
		}
		jc.repo = repo
		return nil
	}
}

func WithAPIClient(apiClient APIClient) Option {
	return func(jc *JiraConnector) error {
		if apiClient == nil {
			return fmt.Errorf("ERROR: apiClient is nil")
		}
		jc.apiClient = apiClient
		return nil
	}
}

func (jc *JiraConnector) GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error) {
	projects, err := jc.apiClient.GetProjects(ctx, limit, page, search)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (jc *JiraConnector) UpdateProject(ctx context.Context, projectKey string) (*Project, error) {
	project, err := jc.apiClient.GetProject(ctx, projectKey)
	if err != nil {
		return nil, err
	}
	err = jc.repo.SaveProject(ctx, *project)
	if err != nil {
		return nil, err
	}
	return project, nil
}
