package db

import (
	"context"
	"errors"
	"foxy/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbpool *pgxpool.Pool

func NewClient() (*pgxpool.Pool, error) {
	if env.FoxyEnvironment.DatabaseUrl == nil {
		return nil, errors.New("no database url set")
	}

	if dbpool != nil {
		return dbpool, nil
	}

	poolConfig, err := pgxpool.ParseConfig(*env.FoxyEnvironment.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	dbpool = pool

	return dbpool, nil
}

func NewConnection() (*pgxpool.Conn, error) {
	pool, err := NewClient()
	if err != nil {
		return nil, err
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Boot() error {
	return RunMigrations()
}
