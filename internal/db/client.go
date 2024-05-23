package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

var dbpool *pgxpool.Pool

func NewClient() (*pgxpool.Pool, error) {
	if dbpool != nil {
		return dbpool, nil
	}

	poolConfig, err := pgxpool.ParseConfig(os.Getenv("DB_URL"))
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}
