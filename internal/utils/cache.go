package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

func GetSourceImage(sourceImageUrl string) (*vips.ImageRef, error) {
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

func GetCachedSource(sid string, key string, sourceImageUrl string, skipCache bool) (*vips.ImageRef, error) {
	if skipCache || os.Getenv("USE_CACHE") != "true" {
		log.Println("Cache is disabled")
		return GetSourceImage(sourceImageUrl)
	}

	sourceFileName := strings.TrimRight(os.Getenv("CACHE_DIR"), "/") + "/" + sid + "/" + strings.TrimLeft(key, "/")
	_, err := os.Stat(sourceFileName)
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

func GetCachedResult(sid string, key string, params []string, format string) (*[]byte, error) {
	if os.Getenv("USE_RENDER_CACHE") != "true" {
		log.Println("Render cache is disabled")
		return nil, nil
	}

	sourceName := filepath.Base(strings.TrimSuffix(key, filepath.Ext(key)))

	sort.Strings(params)
	paramsKey := strings.Join(params, "/")
	hasher := md5.New()
	hasher.Write([]byte(paramsKey))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(os.Getenv("RENDER_CACHE_DIR"), "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format
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
	if os.Getenv("USE_RENDER_CACHE") != "true" {
		log.Println("Render cache is disabled")
		return nil
	}

	sourceName := filepath.Base(strings.TrimSuffix(key, filepath.Ext(key)))

	sort.Strings(params)
	paramsKey := strings.Join(params, "/")
	hasher := md5.New()
	hasher.Write([]byte(paramsKey))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedFileName := strings.TrimRight(os.Getenv("RENDER_CACHE_DIR"), "/") + "/" + sid + "/" + sourceName + "/" + hash + "." + format

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
