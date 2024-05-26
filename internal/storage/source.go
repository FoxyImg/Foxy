package storage

import (
	"errors"
	"fmt"
	"foxy/internal/aws"
	"foxy/internal/config"
	"foxy/internal/env"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/cyphar/filepath-securejoin"
)

func GetSourceImage(config *config.Config, sid string, key string, disableSourceCache bool) (*vips.ImageRef, error) {
	if config.Source.Type == "s3" {
		url, err := aws.GetSignedUrl(*config, key, time.Hour*1)
		if err != nil {
			return nil, err
		}

		return GetCachedSourceFromUrl(sid, key, url, disableSourceCache)
	} else if config.Source.Type == "web" {
		url := *config.Source.WebConfig.Url + "/" + key
		return GetCachedSourceFromUrl(sid, key, url, disableSourceCache)
	} else if config.Source.Type == "local" {

		filePath, err := securejoin.SecureJoin(*config.Source.LocalConfig.Path, "/"+key)
		if err != nil {
			return nil, err
		}

		_, err = os.Stat(filePath)
		if err != nil {
			return nil, err
		}

		img, err := vips.NewImageFromFile(filePath)
		if err != nil {
			log.Println("New Image Error:", err)
			return nil, err
		}

		return img, nil
	} else {
		return nil, errors.New("unknown source type")
	}
}

func GetSourceImageRef(sourceImageUrl string) (*vips.ImageRef, error) {
	response, err := http.Get(sourceImageUrl)
	if err != nil {
		log.Println("HTTP Get Error:", err)
		return nil, err
	}

	sourceImage, imageErr := vips.NewImageFromReader(response.Body)
	if imageErr != nil {
		log.Println("Read Image Error:", imageErr)
		return nil, imageErr
	}

	if sourceImage == nil {
		log.Println("Error: image1 is nil")
		return nil, errors.New("unable to read image")
	}

	return sourceImage, nil
}

func GetCachedSourceFromUrl(sid string, key string, sourceImageUrl string, skipCache bool) (*vips.ImageRef, error) {
	if skipCache || !env.FoxyEnvironment.UseCache || env.FoxyEnvironment.CacheDir == nil {
		log.Println("Cache is disabled")
		return GetSourceImageRef(sourceImageUrl)
	}

	sourceFilePath := "/" + sid + "/" + strings.TrimLeft(key, "/")
	sourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), sourceFilePath)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(sourceFileName)
	if err == nil {
		log.Println("Cache hit")
		img, err := vips.NewImageFromFile(sourceFileName)
		if err != nil {
			log.Println("New Image Error:", err)
			return nil, err
		}

		return img, nil
	}

	sourcePath := filepath.Dir(sourceFileName)
	err = os.MkdirAll(sourcePath, os.ModePerm)
	if err != nil {
		log.Println("MkdirAll Error:", err)
		return nil, err
	}

	out, err := os.Create(sourceFileName)
	if err != nil {
		log.Println("Create Error:", err)
		return nil, err
	}

	resp, err := http.Get(sourceImageUrl)
	if err != nil {
		_ = out.Close()
		_ = os.Remove(sourceFileName)
		log.Println("HTTP Get Error:", err)
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_ = out.Close()
		_ = os.Remove(sourceFileName)

		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	_ = out.Close()
	if err != nil {
		_ = os.Remove(sourceFileName)
		log.Println("Copy Error:", err)
		return nil, err
	}

	img, err := vips.NewImageFromFile(sourceFileName)
	if err != nil {
		log.Println("New Image Error:", err)
		return nil, err
	}

	return img, nil
}
