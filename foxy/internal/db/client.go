package db

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
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

type SourceConfig struct {
	SourceType string  `json:"type"`
	Bucket     *string `json:"bucket"`
	Region     *string `json:"region"`
	Key        *string `json:"key"`
	Secret     *string `json:"secret"`

	RekognitionRegion *string `json:"rekognitionRegion"`
	RekognitionKey    *string `json:"rekognitionKey"`
	RekognitionSecret *string `json:"rekognitionSecret"`
}

func GetSourceConfig(sid string) (*SourceConfig, error) {
	var configJSON string

	redisConfigJSON, err := RedisGet("config:" + sid)
	if err != nil {
		return nil, err
	}

	if redisConfigJSON != nil {
		configJSON = *redisConfigJSON
		log.Println("Using config from redis")
	} else {
		pg, err := NewClient()
		if err != nil {
			return nil, err
		}

		conn, err := pg.Acquire(context.Background())
		if err != nil {
			return nil, err
		}

		res := conn.QueryRow(context.Background(), "SELECT config FROM sources where sid = $1", sid)
		err = res.Scan(&configJSON)
		if err != nil {
			return nil, err
		}

		log.Println("Using config from db")

		RedisSet("config:"+sid, &configJSON, 60*time.Minute)

		conn.Release()
	}

	var result SourceConfig
	err = json.Unmarshal([]byte(configJSON), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
