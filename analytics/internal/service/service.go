package service

import (
	"context"
	"fmt"

	"github.com/sssidkn/JIRA-analyzer/internal/repository"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
)

type Service interface {
	MakeTask(ctx context.Context, task int, key string) (interface{}, error)
}

type service struct {
	repo     repository.Repository
	log      logger.Logger
	handlers map[int]TaskHandler
}

func New(repo repository.Repository, log logger.Logger) *service {
	s := &service{repo: repo, log: log, handlers: make(map[int]TaskHandler)}
	s.handlers[1] = s.makeTaskOne // time in the open state
	s.handlers[2] = s.makeTaskTwo //number of tasks by priority level
	return s
}

type TaskHandler func(ctx context.Context, param string) (interface{}, error)

func (s *service) MakeTask(ctx context.Context, task int, key string) (interface{}, error) {
	if hand, exists := s.handlers[task]; exists {
		return hand(ctx, key)
	}
	return nil, fmt.Errorf("task %d not found", task)
}

func (s *service) makeTaskOne(ctx context.Context, key string) (interface{}, error) {
	issues, err := s.repo.MakeTaskOne(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error making analytical data for task one: %w", err))
		return nil, err
	}
	return issues, nil
}

// other type of priority?
func (s *service) makeTaskTwo(ctx context.Context, key string) (interface{}, error) {
	issues, err := s.repo.MakeTaskTwo(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error making analytical data for task two: %w", err))
		return nil, err
	}
	return issues, nil
}
