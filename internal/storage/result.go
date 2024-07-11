package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/env"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetCachedResult(sourceConfig *config.Config, sid string, key string, params any, format string) (*string, error) {
	renderCacheDir := sourceConfig.RenderCache

	if renderCacheDir == nil && (!env.FoxyEnvironment.UseRenderCache || env.FoxyEnvironment.RenderCacheDir == nil) {
		log.Println("Render cache is disabled")
		return nil, nil
	}

	if renderCacheDir == nil {
		renderCacheDir = env.FoxyEnvironment.RenderCacheDir
	}

	sourceName := strings.TrimSuffix(key, filepath.Ext(key))

	var paramsJSON []byte
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		log.Println("Marshal JSON Error:", err)
		return nil, err
	}

	hasher := md5.New()
	hasher.Write(paramsJSON)
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(*renderCacheDir, "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format
	_, err = os.Stat(hashedFileName)
	if err == nil {
		log.Println("Render cache hit")
		return &hashedFileName, nil
	}

	return nil, nil
}

func SetCachedResult(sourceConfig *config.Config, sid string, key string, params any, format string, data *[]byte) error {
	renderCacheDir := sourceConfig.RenderCache

	if renderCacheDir == nil && (!env.FoxyEnvironment.UseRenderCache || env.FoxyEnvironment.RenderCacheDir == nil) {
		log.Println("Render cache is disabled")
		return nil
	}

	if renderCacheDir == nil {
		renderCacheDir = env.FoxyEnvironment.RenderCacheDir
	}

	sourceName := strings.TrimSuffix(key, filepath.Ext(key))

	var paramsJSON []byte
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		log.Println("Marshal JSON Error:", err)
		return err
	}

	hasher := md5.New()
	hasher.Write(paramsJSON)
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(*renderCacheDir, "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format

	hashedPath := filepath.Dir(hashedFileName)
	err = os.MkdirAll(hashedPath, os.ModePerm)
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
