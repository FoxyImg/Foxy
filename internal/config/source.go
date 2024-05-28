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

type WebConfig struct {
	Url *string `json:"url"`
}

type LocalConfig struct {
	Path *string `json:"path"`
}

type SourceConfig struct {
	Type string `json:"type"`
	S3Config
	WebConfig
	LocalConfig
}

type VisionConfig struct {
	Enabled              bool   `json:"enabled"`
	Type                 string `json:"type"`
	UseSourceCredentials bool   `json:"useSourceCredentials"`
	S3Config
}

type Config struct {
	AppId  *string       `json:"appId"`
	Secret *string       `json:"secret"`
	Source *SourceConfig `json:"source"`
	Vision *VisionConfig `json:"vision"`
}

func GetSourceConfig(sid string) (*Config, error) {
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
		pg, pgErr := db.NewClient()
		if pgErr != nil {
			return nil, pgErr
		}

		conn, acquireErr := pg.Acquire(context.Background())
		if acquireErr != nil {
			return nil, acquireErr
		}

		var configJSON string
		var secret string
		var appId string

		res := conn.QueryRow(context.Background(), "SELECT app_id, key as secret, config FROM sources_view where sid = $1", sid)
		err = res.Scan(&appId, &secret, &configJSON)
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
		result.AppId = &appId

		newJSON, newJSONErr := json.Marshal(result)
		if newJSONErr != nil {
			return nil, newJSONErr
		}

		err = db.RedisSet("config:"+sid, string(newJSON), 60*time.Minute)
		if err != nil {
			log.Println("RedisSet Error:", err)
		}
	}

	return &result, nil
}
