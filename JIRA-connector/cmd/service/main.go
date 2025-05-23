package main

import (
	"context"
	"fmt"
	"jira-connector/internal/config"
	"jira-connector/internal/jira"
	"jira-connector/internal/repository"
	connector "jira-connector/internal/service"
	"jira-connector/internal/transport/grpc/server"
	connectorApi "jira-connector/pkg/api/connector"
	"jira-connector/pkg/db/postgres"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func runGRPCGateway(cfg config.Config) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := connectorApi.RegisterJiraConnectorHandlerFromEndpoint(ctx, mux, fmt.Sprintf("%s:%d", cfg.Host, cfg.PortGRPC), opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.PortHTTP), mux)
}

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	jiraClient := jira.NewClient(cfg.Jira)

	dbPool, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	repo := repository.NewProjectRepository(dbPool)

	jc, err := connector.NewJiraConnector(
		connector.WithAPIClient(jiraClient),
		connector.WithRepository(repo),
	)

	if err != nil {
		panic(err)
	}
	//project, err := jc.UpdateProject(context.Background(), "ATLAS")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("%+v", project)

	grpcServer := grpc.NewServer()
	connectorApi.RegisterJiraConnectorServer(grpcServer, server.NewGRPCServer(jc))

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.PortGRPC))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := runGRPCGateway(*cfg); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}
