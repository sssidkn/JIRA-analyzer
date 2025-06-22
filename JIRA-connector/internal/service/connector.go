package connector

import (
	"context"
	"fmt"
	"jira-connector/internal/models"
	connectorApi "jira-connector/pkg/api/connector"
	"jira-connector/pkg/logger"
	"strings"
	"time"
)

type JiraConnector struct {
	repo      Repository
	apiClient APIClient
	logger    logger.Logger
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

type Issue = models.JiraIssue
type Project = models.JiraProject
type Option = func(*JiraConnector) error

type Repository interface {
	SaveProject(ctx context.Context, project Project) error
	GetProjectInfo(ctx context.Context, projectKey string) (*models.ProjectInfo, error)
}

type APIClient interface {
	UpdateProject(ctx context.Context, projectKey string, lastUpdate time.Time) (*[]models.JiraIssue, error)
	GetProject(ctx context.Context, projectKey string) (*Project, error)
	GetProjects(ctx context.Context, limit, page int, search string) ([]models.ProjectInfo, error)
	GetBaseURL() string
}

func WithLogger(log logger.Logger) Option {
	return func(jc *JiraConnector) error {
		if log == nil {
			log = logger.NewLogrusLogger()
			log.SetLevel(logger.LevelInfo)
		}
		jc.logger = log
		return nil
	}
}

func WithRepository(repo Repository) Option {
	return func(jc *JiraConnector) error {
		if repo == nil {
			return fmt.Errorf("repo is nil")
		}
		jc.repo = repo
		return nil
	}
}

func WithAPIClient(apiClient APIClient) Option {
	return func(jc *JiraConnector) error {
		if apiClient == nil {
			return fmt.Errorf("apiClient is nil")
		}
		jc.apiClient = apiClient
		return nil
	}
}

func (jc *JiraConnector) GetProjects(ctx context.Context, limit, page int, search string) (*connectorApi.GetProjectsResponse, error) {
	projects, err := jc.apiClient.GetProjects(ctx, limit, page, search)
	if err != nil {
		return nil, err
	}
	p := make([]*connectorApi.JiraProject, 0)
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), strings.ToLower(search)) {
			p = append(p, &connectorApi.JiraProject{
				Id:   project.ID,
				Url:  project.Self,
				Key:  project.Key,
				Name: project.Name,
			})
		}
	}
	if limit == 0 {
		return nil, fmt.Errorf("limit cannot be zero")
	}
	startIndex := (page - 1) * limit
	endIndex := page * limit
	if endIndex > len(projects) {
		endIndex = len(projects)
	}
	projectsCount := len(p)
	pageCount := projectsCount / limit
	if projectsCount%limit != 0 {
		pageCount++
	}
	if len(p) != 0 {
		p = p[startIndex:min(endIndex, projectsCount)]
	}
	return &connectorApi.GetProjectsResponse{
		Projects: p,
		PageInfo: &connectorApi.PageInfo{
			PageCount:     int64(pageCount),
			ProjectsCount: int64(projectsCount),
		}}, nil
}

func (jc *JiraConnector) UpdateProject(ctx context.Context, projectKey string) (*Project, error) {
	jc.logger.Debug("Updating project", logger.Field{Key: "project_key", Value: projectKey})
	projectInfo, err := jc.repo.GetProjectInfo(ctx, projectKey)
	if err != nil {
		return nil, err
	}
	updateTime := time.Now()
	var project *Project
	if projectInfo == nil {
		jc.logger.Info("Project not found in DB", logger.Field{Key: "project_key", Value: projectKey})
		jc.logger.Info("Fetching project from JIRA", logger.Field{Key: "project_key", Value: projectKey})
		project, err = jc.apiClient.GetProject(ctx, projectKey)
		if err != nil {
			return nil, err
		}
		project.LastUpdate = updateTime
	} else {
		jc.logger.Info("Project found in DB", logger.Field{Key: "project_key", Value: projectKey})
		jc.logger.Info("Fetching project from JIRA", logger.Field{Key: "project_key", Value: projectKey})
		issues, err := jc.apiClient.UpdateProject(ctx, projectKey, projectInfo.LastUpdate)
		if err != nil {
			return nil, err
		}
		project = &Project{
			ID:         projectInfo.ID,
			Key:        projectKey,
			Name:       projectInfo.Name,
			Self:       jc.apiClient.GetBaseURL() + "/projects/" + projectInfo.Key,
			Issues:     *issues,
			LastUpdate: updateTime,
		}
		if len(*issues) == 0 {
			jc.logger.Info("No new issues found", logger.Field{Key: "project_key", Value: projectKey})
			return project, nil
		}
	}
	jc.logger.Info("Saving project to DB", logger.Field{Key: "project_key", Value: projectKey})
	err = jc.repo.SaveProject(ctx, *project)
	if err != nil {
		jc.logger.Error("Failed to save project to DB", logger.Field{Key: "project_key", Value: projectKey})
		return nil, err
	}
	jc.logger.Info("Project saved to DB", logger.Field{Key: "project_key", Value: projectKey})
	return project, nil
}
