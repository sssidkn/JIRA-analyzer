package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	DbUser     string `yaml:"dbUser"`
	DbPassword string `yaml:"dbPassword"`
	DbHost     string `yaml:"dbHost"`
	DbPort     int    `yaml:"dbPort"`
	DbName     string `yaml:"dbName"`
}

func New(cfg Config) (*pgx.Conn, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		cfg.DbPort,
		cfg.DbName,
	)
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return conn, nil
}
