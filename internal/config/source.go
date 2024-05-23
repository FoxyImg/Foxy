package config

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"log"
	"time"
)

type S3Config struct {
	Key    *string `json:"key"`
	Secret *string `json:"secret"`
	Bucket *string `json:"bucket"`
	Region *string `json:"region"`
}

type SourceConfig struct {
	Type string `json:"type"`
	S3Config
}

type VisionConfig struct {
	Enabled              bool   `json:"enabled"`
	Type                 string `json:"type"`
	UseSourceCredentials bool   `json:"useSourceCredentials"`
	S3Config
}

type Config struct {
	Secret *string       `json:"secret"`
	Source *SourceConfig `json:"source"`
	Vision *VisionConfig `json:"vision"`
}

func GetSourceConfig(sid string) (*Config, error) {
	var configJSON string
	var secret string

	redisConfigJSON, err := db.RedisGet("config:" + sid)
	if err != nil {
		log.Println("RedisGet Error:", err)
		return nil, err
	}

	var result Config
	if redisConfigJSON != nil {
		log.Println("Using config from redis")
		err = json.Unmarshal([]byte(*redisConfigJSON), &result)
		if err != nil {
			return nil, err
		}
	} else {
		pg, err := db.NewClient()
		if err != nil {
			return nil, err
		}

		conn, err := pg.Acquire(context.Background())
		if err != nil {
			return nil, err
		}

		res := conn.QueryRow(context.Background(), "SELECT key as secret, config FROM sources_view where sid = $1", sid)
		err = res.Scan(&secret, &configJSON)
		if err != nil {
			return nil, err
		}

		log.Println("Using config from db")

		conn.Release()

		err = json.Unmarshal([]byte(configJSON), &result)
		if err != nil {
			return nil, err
		}

		result.Secret = &secret

		newJSON, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		err = db.RedisSet("config:"+sid, string(newJSON), 60*time.Minute)
		if err != nil {
			log.Println("RedisSet Error:", err)
		}
	}

	return &result, nil
}
