package service

import (
	"context"
	"fmt"

	"github.com/sssidkn/resources/internal/models"
	"github.com/sssidkn/resources/internal/repository"
	"github.com/sssidkn/resources/pkg/logger"
)

type Service interface {
	GetProjects(ctx context.Context, limit int, offset int) (*models.PaginatedResponse, error)
	GetProject(ctx context.Context, id int) (*models.Response, error)
	DeleteProject(ctx context.Context, id int) error
	GetIssue(ctx context.Context, id int) (*models.Response, error)
	GetIssuesByProject(ctx context.Context, projectId int, limit int, offset int) (*models.PaginatedResponse, error)
	GetHistoryByIssue(ctx context.Context, issueId int) (*models.Response, error)
	GetHistoryByAuthor(ctx context.Context, authorId int) (*models.Response, error)
}

type service struct {
	repo  repository.Repository
	log   logger.Logger
	links *models.ReferencesLinks
}

func New(repo repository.Repository, log logger.Logger, port int) *service {
	return &service{repo: repo, log: log, links: &models.ReferencesLinks{LinkProjects: []models.Link{{fmt.Sprintf("http://localhost:%d/api/v1/projects", port)}},
		LinkIssues: []models.Link{{fmt.Sprintf("http://localhost:%d/api/v1/issues", port)},
			{fmt.Sprintf("http://localhost:%d/api/v1/issues/by-project", port)}},
		LinkHistories: []models.Link{{fmt.Sprintf("http://localhost:%d/api/v1/histories/by-issue", port)},
			{fmt.Sprintf("http://localhost:%d/api/v1/histories/by-author", port)}}}}
}

func (s *service) GetProjects(ctx context.Context, limit int, offset int) (*models.PaginatedResponse, error) {
	projects, total, err := s.repo.GetProjects(ctx, limit, offset)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting projects: %w", err))
		return nil, err
	}

	pageInfo := getPageInfo(total, limit, offset)
	var response models.PaginatedResponse
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = projects
	response.PageInfo = pageInfo
	return &response, nil
}

func (s *service) GetProject(ctx context.Context, id int) (*models.Response, error) {
	project, err := s.repo.GetProject(ctx, id)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting project: %w", err))
		return nil, err
	}

	var response models.Response
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = project
	return &response, nil
}

func (s *service) DeleteProject(ctx context.Context, id int) error {
	err := s.repo.DeleteProject(ctx, id)
	if err != nil {
		s.log.Error(fmt.Errorf("error deleting project: %w", err))
		return err
	}
	return nil
}

func (s *service) GetIssue(ctx context.Context, id int) (*models.Response, error) {
	issue, err := s.repo.GetIssue(ctx, id)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting issue: %w", err))
		return nil, err
	}

	var response models.Response
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = issue
	return &response, nil
}

func (s *service) GetIssuesByProject(ctx context.Context, projectId int, limit int, offset int) (*models.PaginatedResponse, error) {
	issues, total, err := s.repo.GetIssuesByProject(ctx, projectId, limit, offset)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting issues : %w", err))
		return nil, err
	}

	pageInfo := getPageInfo(total, limit, offset)
	var response models.PaginatedResponse
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = issues
	response.PageInfo = pageInfo
	return &response, nil
}

func (s *service) GetHistoryByIssue(ctx context.Context, issueId int) (*models.Response, error) {
	history, err := s.repo.GetHistoryByIssue(ctx, issueId)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting history : %w", err))
		return nil, err
	}

	var response models.Response
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = history
	return &response, nil
}

func (s *service) GetHistoryByAuthor(ctx context.Context, authorId int) (*models.Response, error) {
	history, err := s.repo.GetHistoryByAuthor(ctx, authorId)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting history : %w", err))
		return nil, err
	}

	var response models.Response
	links, err := s.addLink(ctx)
	if err == nil {
		response.Links = links
	}
	response.Data = history
	return &response, nil
}

func (s *service) addLink(ctx context.Context) (models.ReferencesLinks, error) {
	self, ok := ctx.Value("url").(string)
	if ok {
		links := *s.links
		links.LinkSelf = models.Link{URL: self}
		return links, nil
	}
	return models.ReferencesLinks{}, fmt.Errorf("error getting self url")
}

func getPageInfo(total, limit, offset int) models.Pagination {
	pageCount := total / limit
	if total%limit != 0 {
		pageCount++
	}
	currentPage := offset/limit + 1
	return models.Pagination{CurrentPage: currentPage, PageCount: pageCount, Total: total}
}
