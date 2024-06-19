package config

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"foxy/internal/env"
	"log"
	"os"
	"sync"
	"time"
)

type APIKeys struct {
	PhotoRoom *string `json:"photoroom,omitempty"`
	ClipDrop  *string `json:"clipdrop,omitempty"`
}

type S3Config struct {
	Key    *string `json:"key,omitempty"`
	Secret *string `json:"secret,omitempty"`
	Bucket *string `json:"bucket,omitempty"`
	Region *string `json:"region,omitempty"`
}

type WebConfig struct {
	Url *string `json:"url,omitempty"`
}

type LocalConfig struct {
	Path *string `json:"path,omitempty"`
}

type SourceConfig struct {
	Type string `json:"type"`
	S3Config
	WebConfig
	LocalConfig
}

type CompreFaceConfig struct {
	ApiKey *string `json:"apiKey,omitempty"`
	Url    *string `json:"url,omitempty"`
}

type VisionConfig struct {
	Enabled              bool   `json:"enabled"`
	Type                 string `json:"type"`
	UseSourceCredentials bool   `json:"useSourceCredentials"`
	S3Config
	CompreFaceConfig
}

type Config struct {
	AppId   *string       `json:"appId"`
	Secret  *string       `json:"secret"`
	Source  *SourceConfig `json:"source"`
	Vision  *VisionConfig `json:"vision"`
	APIKeys *APIKeys      `json:"keys"`
}

var LoadedSources *map[string]*Config = nil

func LoadSourceConfigFromJSON() error {
	if LoadedSources != nil {
		return nil
	}

	jsonData, err := os.ReadFile(*env.FoxyEnvironment.SourceConfigFile)
	if err != nil {
		log.Println("Read File Error:", err)
		return err
	}

	sources := make(map[string]*Config)
	err = json.Unmarshal(jsonData, &sources)
	if err != nil {
		log.Println("Unmarshal Error:", err)
		return err
	}

	m := &sync.Mutex{}
	m.Lock()
	LoadedSources = &sources
	m.Unlock()

	return nil
}

func GetSourceConfig(sid string) (*Config, error) {
	if env.FoxyEnvironment.Isolated {
		if LoadedSources == nil {
			err := LoadSourceConfigFromJSON()
			if err != nil {
				return nil, err
			}
		}

		sourceConfig := (*LoadedSources)[sid]

		return sourceConfig, nil
	}

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
	} else if env.FoxyEnvironment.DatabaseUrl != nil {
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
	} else {
		return nil, nil
	}

	return &result, nil
}
