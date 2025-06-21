package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `yaml:"Host" env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `yaml:"Port" env:"POSTGRES_PORT" envDefault:"5434"`
	User     string `yaml:"User" env:"POSTGRES_USER" envDefault:"root"`
	Password string `yaml:"Password" env:"POSTGRES_PASSWORD"`
	Database string `yaml:"Database" env:"POSTGRES_DB" envDefault:"BowCompetitions"`
	PoolSize int32  `yaml:"PoolSize" env:"POSTGRES_POOL_SIZE" envDefault:"10"`
}

func New(config Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.MaxConns = config.PoolSize

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create connection pool: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return pool, nil
}
