package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/sssidkn/jira-connector/internal/models"
	connectorApi "github.com/sssidkn/jira-connector/pkg/api/connector"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type Option func(*GRPCServer)

type Service interface {
	UpdateProject(ctx context.Context, projectKey string) (*models.JiraProject, error)
	GetProjects(ctx context.Context, limit, page int, search string) (*connectorApi.GetProjectsResponse, error)
}

type GRPCServer struct {
	connectorApi.UnimplementedJiraConnectorServer
	server  *grpc.Server
	service Service
	wg      *sync.WaitGroup
	logger  *logger.Logger
}

func NewGRPCServer(options ...Option) *GRPCServer {
	srv := &GRPCServer{}
	srv.wg = &sync.WaitGroup{}
	for _, opt := range options {
		opt(srv)
	}
	return srv
}

func WithService(service Service) Option {
	return func(s *GRPCServer) {
		s.service = service
	}
}

func WithLogger(log logger.Logger) Option {
	if log == nil {
		log = logger.NewLogrusLogger()
	}
	return func(s *GRPCServer) {
		s.logger = &log
	}
}

func (s *GRPCServer) UpdateProject(ctx context.Context,
	req *connectorApi.UpdateProjectRequest) (*connectorApi.UpdateProjectResponse, error) {
	project, err := s.service.UpdateProject(ctx, req.GetProjectKey())
	if err != nil {
		return nil, err
	}

	return &connectorApi.UpdateProjectResponse{
		Project: &connectorApi.JiraProject{
			Id:   project.ID,
			Url:  project.Self,
			Key:  project.Key,
			Name: project.Name,
		},
		Success: true,
	}, nil
}

func (s *GRPCServer) GetProjects(ctx context.Context, req *connectorApi.GetProjectsRequest) (*connectorApi.GetProjectsResponse, error) {
	response, err := s.service.GetProjects(ctx, int(req.GetLimit()), int(req.GetPage()), req.GetSearch())
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *GRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(logger.Interceptor(*s.logger)),
	)
	connectorApi.RegisterJiraConnectorServer(s.server, s)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			(*s.logger).Error(fmt.Sprintf("gRPC server failed: %v", err))
		}
	}()

	(*s.logger).Info(fmt.Sprintf("gRPC server started on %s", addr))
	return nil
}

func (s *GRPCServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		s.wg.Wait()
		(*s.logger).Info("gRPC server stopped gracefully")
	}
}
