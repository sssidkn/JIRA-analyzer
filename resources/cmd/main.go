package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/sssidkn/JIRA-analyzer/docs"
	"github.com/sssidkn/JIRA-analyzer/internal/config"
	"github.com/sssidkn/JIRA-analyzer/internal/repository"
	"github.com/sssidkn/JIRA-analyzer/internal/server"
	"github.com/sssidkn/JIRA-analyzer/internal/service"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
	"github.com/sssidkn/JIRA-analyzer/pkg/postgres"
)

// @title Resources Swagger API
// @version 1.0
// @description Swagger API for Golang Project Blueprint.
// @termsOfService http://swagger.io/terms/

// @license.name MIT
// @BasePath /api/v1

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	lg, err := logger.New("logs")
	if err != nil {
		log.Fatal(err)
	}
	lg.Debug("logger was created")

	cfg, err := config.New("./config/config.yaml")
	if err != nil {
		lg.Fatal(err)
	}

	errs := config.MissingSetting(cfg)
	if len(errs) > 0 {
		lg.Fatal(errs)
	}

	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Debug("successful connection to postgres")

	rp := repository.New(pg)
	sv := service.New(rp, *lg, cfg.Port)
	server := server.New(sv, lg, cfg.ResourceTimeout)

	go func() {
		lg.Info("Server is listening on port:" + strconv.Itoa(cfg.Port))
		if err = server.Run(cfg.Port); err != nil {
			stop()
		}
	}()

	<-ctx.Done()
	lg.Info("Shutting down server...")
	server.Shutdown(ctx)
	lg.Info("Server shut down")
	defer pg.Close()
}
