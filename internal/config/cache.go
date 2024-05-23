package config

import (
	"context"
	"encoding/json"
	"foxy/internal/db"
	"log"
	"time"
)

type configCacheEntry struct {
	config  *Config
	expires time.Time
}

var configCache map[string]*configCacheEntry

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

func GetConfigCache(accessKey string) (*Config, error) {
	if configCache == nil {
		configCache = make(map[string]*configCacheEntry)
	}

	result := configCache[accessKey]
	if result == nil || result.expires.Before(time.Now()) {
		log.Println("Config cache miss, fetching from db")
		sourceConfig, err := GetSourceConfig(accessKey)
		if err != nil {
			return nil, err
		}

		result = &configCacheEntry{
			config:  sourceConfig,
			expires: time.Now().Add(time.Minute * 1),
		}

		configCache[accessKey] = result
	} else {
		log.Println("Using config from cache")
	}

	return result.config, nil
}
