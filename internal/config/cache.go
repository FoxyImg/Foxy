package config

import (
	"log"
	"os"
	"sync"
	"time"
)

type configCacheEntry struct {
	config  *Config
	expires time.Time
}

var configCache map[string]*configCacheEntry

func GetSourceConfigFromCache(accessKey string) (*Config, error) {
	if os.Getenv("USE_SOURCE_CONFIG_CACHE") == "true" {
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
				expires: time.Now().Add(time.Minute * 15),
			}

			var mutex = &sync.Mutex{}
			mutex.Lock()
			configCache[accessKey] = result
			mutex.Unlock()
		} else {
			log.Println("Using config from cache")
		}
		return result.config, nil
	} else {
		log.Println("Source config cache is disabled")
		return GetSourceConfig(accessKey)
	}
}
