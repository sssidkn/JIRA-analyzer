package server

import (
	"context"
	"errors"
	"fmt"
	connector "github.com/sssidkn/jira-connector/internal/service"
	connectorApi "github.com/sssidkn/jira-connector/pkg/api/connector"
	"github.com/sssidkn/jira-connector/pkg/logger"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Option func(*HTTPServer)

type HTTPServer struct {
	server   *http.Server
	service  *connector.JiraConnector
	wg       *sync.WaitGroup
	logger   *logger.Logger
	grpcAddr string
}

func NewHTTPServer(options ...Option) *HTTPServer {
	srv := &HTTPServer{
		wg: &sync.WaitGroup{},
	}
	for _, opt := range options {
		opt(srv)
	}
	return srv
}

func WithService(service *connector.JiraConnector) Option {
	return func(s *HTTPServer) {
		s.service = service
	}
}

func WithLogger(log logger.Logger) Option {
	if log == nil {
		log = logger.NewLogrusLogger()
	}
	return func(s *HTTPServer) {
		s.logger = &log
	}
}

func WithGRPCAddress(addr string) Option {
	return func(s *HTTPServer) {
		s.grpcAddr = addr
	}
}

func (s *HTTPServer) Start(addr string) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	ctx := context.Background()
	err := connectorApi.RegisterJiraConnectorHandlerFromEndpoint(
		ctx,
		mux,
		s.grpcAddr,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register HTTP gateway: %w", err)
	}

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			(*s.logger).Error(fmt.Sprintf("HTTP server failed: %v", err))
		}
	}()

	(*s.logger).Info(fmt.Sprintf("HTTP gateway started on %s", addr))
	return nil
}

func (s *HTTPServer) Stop() {
	if s.server != nil {
		if err := s.server.Shutdown(context.Background()); err != nil {
			(*s.logger).Error(fmt.Sprintf("HTTP server shutdown error: %v", err))
		}
		s.wg.Wait()
		(*s.logger).Info("HTTP server stopped gracefully")
	}
}
