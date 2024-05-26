package db

import (
	"context"
	"foxy/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbpool *pgxpool.Pool

func NewClient() (*pgxpool.Pool, error) {
	if dbpool != nil {
		return dbpool, nil
	}

	poolConfig, err := pgxpool.ParseConfig(env.FoxyEnvironment.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}
