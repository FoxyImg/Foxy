package storage

import (
	"errors"
	"foxy/internal/aws"
	"foxy/internal/config"
	"foxy/internal/env"
	securejoin "github.com/cyphar/filepath-securejoin"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func GetSourceText(config *config.Config, sourceId string, key string, disableSourceCache bool) (*[]byte, error) {
	if config.Source.Type == "s3" {
		url, err := aws.GetSignedUrl(*config, key, time.Hour*1)
		if err != nil {
			return nil, err
		}

		return GetCachedSourceTextFromUrl(sourceId, key, url, disableSourceCache)
	} else if config.Source.Type == "web" {
		url := *config.Source.WebConfig.Url + "/" + key
		return GetCachedSourceTextFromUrl(sourceId, key, url, disableSourceCache)
	} else if config.Source.Type == "local" {

		filePath, err := securejoin.SecureJoin(*config.Source.LocalConfig.Path, "/"+key)
		if err != nil {
			return nil, err
		}

		_, err = os.Stat(filePath)
		if err != nil {
			return nil, err
		}

		sourceText, err := os.ReadFile(filePath)
		if err != nil {
			log.Println("Read File Error:", err)
			return nil, err
		}

		return &sourceText, nil
	} else {
		return nil, errors.New("unknown source type")
	}
}

func GetUncachedSourceText(sourceImageUrl string) (*[]byte, error) {
	response, err := http.Get(sourceImageUrl)
	if err != nil {
		log.Println("HTTP Get Error:", err)
		return nil, err
	}

	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Read Body Error:", err)
		return nil, err
	}

	return &res, nil
}

func GetCachedSourceTextFromUrl(sourceId string, key string, sourceImageUrl string, skipCache bool) (*[]byte, error) {
	if skipCache || !env.FoxyEnvironment.UseCache || env.FoxyEnvironment.CacheDir == nil {
		log.Println("Cache is disabled")
		return GetUncachedSourceText(sourceImageUrl)
	}

	sourceFileName, err := GetCachedPathFromUrl(sourceId, key, sourceImageUrl, skipCache)
	if err != nil {
		return nil, err
	}

	sourceText, err := os.ReadFile(*sourceFileName)
	if err != nil {
		log.Println("Read File Error:", err)
		return nil, err
	}

	return &sourceText, nil
}
