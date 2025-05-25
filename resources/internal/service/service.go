package service

import (
	"context"
	"fmt"

	"github.com/sssidkn/JIRA-analyzer/internal/models"
	"github.com/sssidkn/JIRA-analyzer/internal/repository"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
)

type Service interface {
	GetProjects(ctx context.Context, limit int, offset int) (*models.Response, error)
	GetProject(ctx context.Context, id int) (*models.Response, error)
	DeleteProject(ctx context.Context, id int) error
}

type service struct {
	repo repository.Repository
	log  logger.Logger
}

func New(repo repository.Repository, log logger.Logger) *service {
	return &service{repo: repo, log: log}
}

func (s *service) GetProjects(ctx context.Context, limit int, offset int) (*models.Response, error) {
	projects, err := s.repo.GetProjects(ctx, limit, offset)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting projects: %w", err))
		return nil, err
	}
	return &models.Response{Data: projects}, nil
}

func (s *service) GetProject(ctx context.Context, id int) (*models.Response, error) {
	project, err := s.repo.GetProject(ctx, id)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting project: %w", err))
		return nil, err
	}
	return &models.Response{Data: project}, nil
}

func (s *service) DeleteProject(ctx context.Context, id int) error {
	err := s.repo.DeleteProject(ctx, id)
	if err != nil {
		s.log.Error(fmt.Errorf("error deleting project: %w", err))
		return err
	}
	return nil
}
