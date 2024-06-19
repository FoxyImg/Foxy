package photoroom

import (
	"bytes"
	"fmt"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/utils"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetBackgroundMask(config *config.Config, sourceId string, key string, sourceImage *vips.ImageRef, skipCache bool) (*vips.ImageRef, error) {
	sourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(key, "/") + ".mask.png"
	sourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), sourceFilePath)
	if !skipCache {
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(sourceFileName)
		if err == nil {
			log.Println("Mask cache hit")
			maskImg, maskImgErr := vips.NewImageFromFile(sourceFileName)
			if maskImgErr != nil {
				log.Println("Read File Error:", err)
				return nil, err
			}

			return maskImg, nil
		}
	}

	if config.APIKeys == nil || config.APIKeys.PhotoRoom == nil {
		return nil, nil
	}

	var pngData []byte
	if sourceImage.Width() > 4000 || sourceImage.Height() > 4000 {
		sourceCopy, err := sourceImage.Copy()
		if err != nil {
			return nil, err
		}

		sc := 4000.0 / float64(utils.Max(sourceCopy.Width(), sourceCopy.Height()))
		_ = sourceCopy.Resize(sc, vips.KernelLanczos3)
		pngData, _, err = sourceCopy.ExportPng(nil)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		pngData, _, err = sourceImage.ExportPng(nil)
		if err != nil {
			return nil, err
		}
	}

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, err := writer.CreateFormFile("image_file", "image.png")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(file, bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://sdk.photoroom.com/v1/segment", payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-api-key", *config.APIKeys.PhotoRoom)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", res.Status)
	}

	defer res.Body.Close()

	resImageData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	maskImg, err := vips.NewImageFromBuffer(resImageData)
	if err != nil {
		return nil, err
	}

	if maskImg.Bands() == 4 {
		alpha, alphaErr := maskImg.ExtractBandToImage(3, 1)
		if alphaErr != nil {
			return nil, alphaErr
		}

		maskImg = alpha
	} else {
		alpha, alphaErr := maskImg.ExtractBandToImage(0, 1)
		if alphaErr != nil {
			return nil, alphaErr
		}

		maskImg = alpha
	}

	if maskImg.Width() != sourceImage.Width() || maskImg.Height() != sourceImage.Height() {
		sz := float64(utils.Max(sourceImage.Width(), sourceImage.Height())) / float64(utils.Max(maskImg.Width(), maskImg.Height()))
		_ = maskImg.Resize(sz, vips.KernelLanczos3)
	}

	pngData, _, err = maskImg.ExportPng(nil)
	if err != nil {
		return maskImg, err
	}

	sourcePath := filepath.Dir(sourceFileName)
	err = os.MkdirAll(sourcePath, os.ModePerm)
	if err != nil {
		log.Println("MkdirAll Error:", err)
		return maskImg, err
	}

	err = os.WriteFile(sourceFileName, pngData, os.ModePerm)
	if err != nil {
		return maskImg, err
	}

	return maskImg, nil
}
