package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sssidkn/jira-connector/internal/config"
	"github.com/sssidkn/jira-connector/internal/jira"
	"github.com/sssidkn/jira-connector/internal/repository"
	connector "github.com/sssidkn/jira-connector/internal/service"
	grpcSrv "github.com/sssidkn/jira-connector/internal/transport/grpc/server"
	httpSrv "github.com/sssidkn/jira-connector/internal/transport/http/server"
	"github.com/sssidkn/jira-connector/pkg/db/postgres"
	"github.com/sssidkn/jira-connector/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	var log logger.Logger = logger.NewLogrusLogger()
	log.SetLevel(cfg.LogLevel)

	log.Info("Initializing jira client...")
	jiraClient := jira.NewClient(
		jira.WithConfig(cfg.Jira),
		jira.WithLogger(log),
		jira.WithMaxDelay(cfg.Jira.MaxDelay),
		jira.WithStartDelay(cfg.Jira.StartDelay),
	)
	log.Info("Jira client initialized")

	log.Info("Initializing db connection...")
	dbPool, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	repo := repository.NewProjectRepository(dbPool)

	repo.SetLogger(log)
	log.Info("DB connection initialized")

	jc, err := connector.NewJiraConnector(
		connector.WithAPIClient(jiraClient),
		connector.WithRepository(repo),
		connector.WithLogger(log),
	)

	if err != nil {
		panic(err)
	}

	grpcServer := grpcSrv.NewGRPCServer(
		grpcSrv.WithService(jc),
		grpcSrv.WithLogger(log),
	)

	err = grpcServer.Start(fmt.Sprintf("%s:%d", cfg.Host, cfg.PortGRPC))
	if err != nil {
		panic(err)
	}
	defer grpcServer.Stop()

	httpServer := httpSrv.NewHTTPServer(
		httpSrv.WithService(jc),
		httpSrv.WithLogger(log),
		httpSrv.WithGRPCAddress(fmt.Sprintf("%s:%d", cfg.Host, cfg.PortGRPC)),
	)
	err = httpServer.Start(fmt.Sprintf("%s:%d", cfg.Host, cfg.PortHTTP))
	if err != nil {
		panic(err)
	}
	defer httpServer.Stop()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Received termination signal, shutting down...")
		cancel()
	case <-ctx.Done():
		log.Info("Context cancelled, shutting down...")
	}
}
