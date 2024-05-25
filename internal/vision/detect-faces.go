package vision

import (
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/metadata"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func DetectFaces(sourceConfig config.Config, sid string, key string, sourceImage *vips.ImageRef, skipCache bool) (*metadata.Metadata, error) {
	metaFilePath := "/" + sid + "/" + strings.TrimLeft(key, "/") + ".json"
	metaFileName, err := securejoin.SecureJoin(strings.TrimRight(os.Getenv("CACHE_DIR"), "/"), metaFilePath)
	if err != nil {
		return nil, err
	}

	if !skipCache && os.Getenv("USE_CACHE") == "true" {
		_, err := os.Stat(metaFileName)
		if err == nil {
			jsonData, err := os.ReadFile(metaFileName)
			if err == nil {
				meta := metadata.Metadata{}
				jsonErr := json.Unmarshal(jsonData, &meta)
				if jsonErr == nil {
					log.Println("Metadata cache hit")
					return &meta, nil
				} else {
					log.Println("Metadata json parse error", jsonErr)
				}
			} else {
				log.Println("Error reading metadata json", err)
			}
		}
	} else {
		log.Println("Skipping meta cache")
	}

	if !sourceConfig.Vision.Enabled {
		return nil, nil
	}

	var meta *metadata.Metadata
	if sourceConfig.Vision.Type == "rekognition" {
		m, err := RekognitionDetectFaces(sourceConfig, sid, key, sourceImage)
		if err != nil {
			return nil, err
		}

		meta = m
	} else {
		return nil, nil
	}

	if meta != nil {
		if os.Getenv("USE_CACHE") == "true" {
			metaFilePath := filepath.Dir(metaFileName)
			err := os.MkdirAll(metaFilePath, os.ModePerm)
			if err != nil {
				log.Println("MkdirAll Error:", err)
				return meta, nil
			}

			jsonData, err := json.Marshal(*meta)
			if err != nil {
				log.Println("Marshal Error:", err)
				return meta, nil
			}

			_ = os.WriteFile(metaFileName, jsonData, 0644)
		}

		return meta, nil
	}

	return nil, nil
}
