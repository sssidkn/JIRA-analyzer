package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sssidkn/JIRA-analyzer/internal/repository"
	"github.com/sssidkn/JIRA-analyzer/pkg/api/connectorApi"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
)

type Service interface {
	MakeTask(ctx context.Context, task int, key string) (interface{}, error)
	GetTask(ctx context.Context, task int, key string) (interface{}, error)
	DeleteTasks(ctx context.Context, key string) (bool, error)
	IsAnalyzed(ctx context.Context, key string) (bool, error)
	Compare(ctx context.Context, task int, keys string) (interface{}, error)
}

type service struct {
	repo     repository.Repository
	log      logger.Logger
	handlers map[int]TaskHandler
	client   connectorApi.JiraConnectorClient
}

func New(repo repository.Repository, log logger.Logger, client connectorApi.JiraConnectorClient) *service {
	s := &service{repo: repo, log: log, handlers: make(map[int]TaskHandler), client: client}
	s.handlers[1] = s.makeTaskOne // time in the open state
	s.handlers[2] = s.makeTaskTwo //number of tasks by priority level
	s.handlers[3] = s.getTaskOne
	s.handlers[4] = s.getTaskTwo
	s.handlers[5] = s.compareTaskOne
	s.handlers[6] = s.compareTaskTwo
	return s
}

type TaskHandler func(ctx context.Context, param string) (interface{}, error)

func (s *service) MakeTask(ctx context.Context, task int, key string) (interface{}, error) {
	response, err := s.client.UpdateProject(ctx, &connectorApi.UpdateProjectRequest{ProjectKey: key})
	if err != nil {
		s.log.Error(fmt.Errorf("failed to update project %s: %w", key, err))
		return nil, err
	}
	if response.Success {
		s.log.Info(fmt.Sprintf("updated project %s", key))
	}
	if hand, exists := s.handlers[task]; exists {
		return hand(ctx, key)
	}
	return nil, fmt.Errorf("task %d not found", task)
}

func (s *service) makeTaskOne(ctx context.Context, key string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	issues, err := s.repo.MakeTaskOne(ctx, key)
	if err != nil {
		if !errors.Is(err, repository.ErrAlreadyExist) {
			s.log.Error(fmt.Errorf("error making analytical data for task one: %w", err))
			return nil, err
		}
		s.log.Info(fmt.Sprintf("data for task 1 for project %s already exists", key))
		issues, err = s.repo.GetTaskOne(ctx, key)
		if err != nil {
			s.log.Error(fmt.Errorf("error getting task 1 for project %s: %w", key, err))
			return nil, err
		}
	}
	return issues, nil
}

// other type of priority?
func (s *service) makeTaskTwo(ctx context.Context, key string) (interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	issues, err := s.repo.MakeTaskTwo(ctx, key)
	if err != nil {
		if !errors.Is(err, repository.ErrAlreadyExist) {
			s.log.Error(fmt.Errorf("error making analytical data for task one: %w", err))
			return nil, err
		}
		s.log.Info(fmt.Sprintf("data for task 2 for project %s already exists", key))
		issues, err = s.repo.GetTaskTwo(ctx, key)
		if err != nil {
			s.log.Error(fmt.Errorf("error getting task 2 for project %s: %w", key, err))
			return nil, err
		}
	}
	return issues, nil
}

func (s *service) GetTask(ctx context.Context, task int, key string) (interface{}, error) {
	if hand, exists := s.handlers[task+2]; exists {
		return hand(ctx, key)
	}
	return nil, fmt.Errorf("task %d not found", task)
}

func (s *service) getTaskOne(ctx context.Context, key string) (interface{}, error) {
	issues, err := s.repo.GetTaskOne(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting task 1 for project %s: %w", key, err))
		return nil, err
	}
	return issues, nil
}

func (s *service) getTaskTwo(ctx context.Context, key string) (interface{}, error) {
	issues, err := s.repo.GetTaskTwo(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error getting task 2 for project %s: %w", key, err))
		return nil, err
	}
	return issues, nil
}

func (s *service) DeleteTasks(ctx context.Context, key string) (bool, error) {
	ok, err := s.repo.DeleteTasks(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error deleting tasks for project %s: %w", key, err))
		return false, err
	}
	if !ok {
		s.log.Info(fmt.Sprintf("no task data to delete for the project %s", key))
	}
	return ok, nil
}

func (s *service) IsAnalyzed(ctx context.Context, key string) (bool, error) {
	isAnalyzed, err := s.repo.IsAnalyzed(ctx, key)
	if err != nil {
		s.log.Error(fmt.Errorf("error checking if analytical data for project %s: %w", key, err))
		return false, err
	}
	return isAnalyzed, nil
}

func (s *service) Compare(ctx context.Context, task int, keys string) (interface{}, error) {
	if hand, exists := s.handlers[task+4]; exists {
		return hand(ctx, keys)
	}
	return nil, fmt.Errorf("task %d not found", task)
}

func (s *service) compareTaskOne(ctx context.Context, keys string) (interface{}, error) {
	keySlice := strings.Split(keys, ",")
	comparisons, err := s.repo.CompareTaskOne(ctx, &keySlice)
	if err != nil {
		s.log.Error(fmt.Errorf("error comparing task 1 for projects %s: %w", keys, err))
		return nil, err
	}
	return comparisons, nil
}

func (s *service) compareTaskTwo(ctx context.Context, keys string) (interface{}, error) {
	keySlice := strings.Split(keys, ",")
	comparisons, err := s.repo.CompareTaskTwo(ctx, &keySlice)
	if err != nil {
		s.log.Error(fmt.Errorf("error comparing task 2 for projects %s: %w", keys, err))
		return nil, err
	}
	return comparisons, nil
}
