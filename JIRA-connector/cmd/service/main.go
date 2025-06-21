package main

import (
	"context"
	"fmt"
	"jira-connector/internal/config"
	"jira-connector/internal/jira"
	"jira-connector/internal/repository"
	connector "jira-connector/internal/service"
	grpcSrv "jira-connector/internal/transport/grpc/server"
	httpSrv "jira-connector/internal/transport/http/server"
	"jira-connector/pkg/db/postgres"
	"jira-connector/pkg/logger"
	"os"
	"os/signal"
	"syscall"
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

	jiraClient := jira.NewClient(
		jira.WithConfig(cfg.Jira),
		jira.WithLogger(log),
	)

	dbPool, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	repo := repository.NewProjectRepository(dbPool)
	repo.SetLogger(log)

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
