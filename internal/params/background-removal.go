package params

import (
	"foxy/internal/apis/clipdrop"
	"foxy/internal/apis/photoroom"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/onnx"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type BackgroundRemovalParams struct {
	Enabled            *bool   `json:"enabled,omitempty"`
	Mode               *string `json:"mode,omitempty"`
	BackgroundImageKey *string `json:"imageKey,omitempty"`
	BackgroundColor    *string `json:"color,omitempty"`
}

func (*BackgroundRemovalParams) Params() []string {
	return []string{"bgr"}
}

func (opts *BackgroundRemovalParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	switch options[0] {
	case "c":
		if len(options) == 3 {
			opts.Enabled = utils.Ptr(true)
			opts.Mode = utils.Ptr(options[1])
			opts.BackgroundColor = &options[2]
		}
	case "img":
		if len(options) == 3 {
			opts.Enabled = utils.Ptr(true)
			opts.Mode = utils.Ptr(options[1])
			opts.BackgroundImageKey = DecodeBase64StringVal(options[2:])
		}
	}

	return
}

func getMLBackgroundMask(foreground bool, config *config.Config, sourceId string, key string, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sourceFilePath := "/" + sourceId + "/" + strings.TrimLeft(key, "/")
	if foreground {
		sourceFilePath += ".foreground.png"
	} else {
		sourceFilePath += ".human.png"
	}

	sourceFileName, err := securejoin.SecureJoin(strings.TrimRight(*env.FoxyEnvironment.CacheDir, "/"), sourceFilePath)
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

	var maskImg *vips.ImageRef
	if foreground {
		maskImg, err = onnx.GenericBackgroundRemoval.ProcessImage(sourceImage)
	} else {
		maskImg, err = onnx.HumanBackgroundRemoval.ProcessImage(sourceImage)
	}
	if err != nil {
		return nil, err
	}

	pngData, _, err := maskImg.ExportPng(nil)
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

func (opts *BackgroundRemovalParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.Mode == nil || opts.Enabled == nil || !*opts.Enabled {
		return sourceImage, nil
	}

	var maskImg *vips.ImageRef
	var err error
	if *opts.Mode == "photoroom" && config.APIKeys != nil && config.APIKeys.PhotoRoom != nil {
		maskImg, err = photoroom.GetBackgroundMask(config, sourceId, sourceKey, sourceImage, false)
		if err != nil {
			return sourceImage, err
		}
	} else if *opts.Mode == "clipdrop" && config.APIKeys != nil && config.APIKeys.ClipDrop != nil {
		maskImg, err = clipdrop.GetBackgroundMask(config, sourceId, sourceKey, sourceImage, false)
		if err != nil {
			return sourceImage, err
		}
	} else if *opts.Mode == "fg" {
		maskImg, err = getMLBackgroundMask(true, config, sourceId, sourceKey, sourceImage)
		if err != nil {
			return sourceImage, err
		}
	} else if *opts.Mode == "person" {
		maskImg, err = getMLBackgroundMask(false, config, sourceId, sourceKey, sourceImage)
		if err != nil {
			return sourceImage, err
		}
	}

	if maskImg == nil {
		return sourceImage, nil
	}

	if sourceImage.HasAlpha() {
		_ = sourceImage.ExtractBand(3, sourceImage.Bands()-3)
	}

	_ = sourceImage.BandJoin(maskImg)
	maskImg.Close()

	if opts.BackgroundImageKey != nil && *opts.BackgroundImageKey != "" {
		oimg, err := storage.GetSourceImage(config, sourceId, *opts.BackgroundImageKey, params.Debug.DisableSourceCache)
		if err != nil {
			return sourceImage, err
		}

		log.Printf("oimg: %d x %d, source: %d x %d", oimg.Width(), oimg.Height(), sourceImage.Width(), sourceImage.Height())

		hs := float64(sourceImage.Width()) / float64(oimg.Width())
		vs := float64(sourceImage.Height()) / float64(oimg.Height())
		sz := math.Max(hs, vs)
		_ = oimg.Resize(sz, vips.KernelLanczos3)
		log.Printf("oimg: %d x %d, source: %d x %d", oimg.Width(), oimg.Height(), sourceImage.Width(), sourceImage.Height())
		err = oimg.Crop(
			(oimg.Width()/2)-(sourceImage.Width()/2),
			(oimg.Height()/2)-(sourceImage.Height()/2),
			sourceImage.Width(),
			sourceImage.Height(),
		)
		if err != nil {
			log.Println("Crop Error:", err)
		}

		if !oimg.HasAlpha() {
			_ = oimg.AddAlpha()
		}

		_ = oimg.Composite(sourceImage, vips.BlendModeOver, 0, 0)
		return oimg, nil
	} else if opts.BackgroundColor != nil && *opts.BackgroundColor != "" {
		backgroundColor, bgColorErr := ParseHexColor(utils.IfNil(opts.BackgroundColor, "00000000"))
		if bgColorErr == nil {
			colorCopy, err := sourceImage.Copy()
			if err != nil {
				return sourceImage, err
			}

			_ = colorCopy.Linear([]float64{0, 0, 0, 0}, []float64{float64(backgroundColor.R), float64(backgroundColor.G), float64(backgroundColor.B), float64(backgroundColor.A)})
			_ = colorCopy.Composite(sourceImage, vips.BlendModeOver, 0, 0)
			return colorCopy, nil
		}
	}

	return sourceImage, nil
}
