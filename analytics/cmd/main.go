package main

import (
	"log"

	"github.com/sssidkn/JIRA-analyzer/internal/config"
	"github.com/sssidkn/JIRA-analyzer/pkg/logger"
	"github.com/sssidkn/JIRA-analyzer/pkg/postgres"
)

func main() {
	cfg, err := config.New("./config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	errs := config.MissingSetting(cfg)
	if len(errs) > 0 {
		log.Fatal(errs)
	}

	lg, err := logger.New("logs")
	if err != nil {
		log.Fatal(err)
	}
	lg.Debug("logger was created")

	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Debug("successful connection to postgres")
}
