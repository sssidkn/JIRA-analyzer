package config

import (
	"fmt"
	"jira-connector/internal/jira"
	"jira-connector/pkg/db/postgres"
	"jira-connector/pkg/logger"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	debug           = "DEBUG"
	production      = "PRODUCTION"
	debugConfigPath = "./config/config.yaml"
)

type Config struct {
	Jira     jira.Config     `yaml:"Jira"`
	Postgres postgres.Config `yaml:"Postgres"`
	PortHTTP uint            `yaml:"PortHTTP" env:"PORT_HTTP"`
	PortGRPC uint            `yaml:"PortGRPC" env:"PORT_GRPC"`
	Host     string          `yaml:"Host" env:"HOST" envDefault:"0.0.0.0"`
	LogLevel logger.Level
}

func New() (*Config, error) {
	cfg := &Config{}
	env := os.Getenv("ENV")
	switch env {
	case debug, "":
		err := cleanenv.ReadConfig(debugConfigPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
		cfg.LogLevel = logger.LevelDebug
		return cfg, nil
	case production:
		err := cleanenv.ReadEnv(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
		cfg.LogLevel = logger.LevelInfo
		return cfg, nil
	default:
		return nil, fmt.Errorf("unknown env: %s", env)
	}
}
