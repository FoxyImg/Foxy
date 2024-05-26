package storage

import (
	"crypto/md5"
	"encoding/hex"
	"foxy/internal/env"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetCachedResult(sid string, key string, params []string, format string) (*[]byte, error) {
	if !env.FoxyEnvironment.UseRenderCache || env.FoxyEnvironment.RenderCacheDir == nil {
		log.Println("Render cache is disabled")
		return nil, nil
	}

	sourceName := filepath.Base(strings.TrimSuffix(key, filepath.Ext(key)))

	sort.Strings(params)
	paramsKey := strings.Join(params, "/")
	hasher := md5.New()
	hasher.Write([]byte(paramsKey))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(*env.FoxyEnvironment.RenderCacheDir, "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format
	_, err := os.Stat(hashedFileName)
	if err == nil {
		log.Println("Render cache hit")
		data, err := os.ReadFile(hashedFileName)
		if err != nil {
			log.Println("Read File Error:", err)
			return nil, err
		}

		return &data, nil
	}

	return nil, nil
}

func SetCachedResult(sid string, key string, params []string, format string, data *[]byte) error {
	if !env.FoxyEnvironment.UseRenderCache || env.FoxyEnvironment.RenderCacheDir == nil {
		log.Println("Render cache is disabled")
		return nil
	}

	sourceName := filepath.Base(strings.TrimSuffix(key, filepath.Ext(key)))

	sort.Strings(params)
	paramsKey := strings.Join(params, "/")
	hasher := md5.New()
	hasher.Write([]byte(paramsKey))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(*env.FoxyEnvironment.RenderCacheDir, "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format

	hashedPath := filepath.Dir(hashedFileName)
	err := os.MkdirAll(hashedPath, os.ModePerm)
	if err != nil {
		log.Println("MkdirAll Error:", err)
		return err
	}

	err = os.WriteFile(hashedFileName, *data, 0644)
	if err != nil {
		log.Println("Write File Error:", err)
		return err
	}

	return nil
}
