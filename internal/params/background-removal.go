package params

import (
	"foxy/internal/apis/clipdrop"
	"foxy/internal/apis/photoroom"
	"foxy/internal/config"
	"foxy/internal/onnx"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
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
		maskImg, err = onnx.GenericBackgroundRemoval.ProcessImage(sourceImage)
		if err != nil {
			return sourceImage, err
		}
	} else if *opts.Mode == "person" {
		maskImg, err = onnx.HumanBackgroundRemoval.ProcessImage(sourceImage)
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
		sourceImage = oimg
	} else if opts.BackgroundColor != nil && *opts.BackgroundColor != "" {
		backgroundColor, bgColorErr := ParseHexColor(utils.IfNil(params.Background.Color, "00000000"))
		if bgColorErr == nil {
			_ = sourceImage.Flatten(&vips.Color{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B})
		}
	}

	return sourceImage, nil
}
