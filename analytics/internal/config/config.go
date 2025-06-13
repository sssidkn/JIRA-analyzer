package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sssidkn/JIRA-analyzer/pkg/postgres"
)

type Config struct {
	Port             int             `yaml:"port" default:"8084"`
	AnalyticsTimeout time.Duration   `yaml:"analyticsTimeout" default:"15s"`
	GrpcServer       string          `yaml:"grpcServer"`
	Postgres         postgres.Config `yaml:"postgres"`
}

func New(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}
	return &cfg, nil
}

func MissingSetting(cfg *Config) []error {
	var errs []error
	if cfg.Postgres.DbUser == "" {
		errs = append(errs, fmt.Errorf("The dbUser is not configured\n"))
	}
	if cfg.Postgres.DbPassword == "" {
		errs = append(errs, fmt.Errorf("The dbPassword is not configured\n"))
	}
	if cfg.Postgres.DbHost == "" {
		errs = append(errs, fmt.Errorf("The dbHost is not configured\n"))
	}
	if cfg.Postgres.DbPort == 0 {
		errs = append(errs, fmt.Errorf("The dbPort is not configured\n"))
	}
	if cfg.Postgres.DbName == "" {
		errs = append(errs, fmt.Errorf("The dbName is not configured\n"))
	}
	return errs
}
