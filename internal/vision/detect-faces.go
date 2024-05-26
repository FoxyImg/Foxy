package vision

import (
	"encoding/json"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/metadata"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func DetectFaces(sourceConfig config.Config, sid string, key string, sourceImage *vips.ImageRef, skipCache bool) (*metadata.Metadata, error) {
	var metaFilePath *string = nil
	var metaFileName *string = nil

	if env.FoxyEnvironment.CacheDir != nil {
		mfp := "/" + sid + "/" + strings.TrimLeft(key, "/") + ".json"
		mfn, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), mfp)
		if err != nil {
			return nil, err
		}

		metaFilePath = &mfp
		metaFileName = &mfn
	}

	if !skipCache && env.FoxyEnvironment.UseCache && env.FoxyEnvironment.CacheDir != nil && metaFilePath != nil && metaFileName != nil {
		_, err := os.Stat(*metaFileName)
		if err == nil {
			jsonData, err := os.ReadFile(*metaFileName)
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
		if env.FoxyEnvironment.UseCache && env.FoxyEnvironment.CacheDir != nil && metaFilePath != nil && metaFileName != nil {
			metaFilePath := filepath.Dir(*metaFileName)
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

			_ = os.WriteFile(*metaFileName, jsonData, 0644)
		}

		return meta, nil
	}

	return nil, nil
}
