package params

import "C"
import (
	"foxy/internal/config"
	"foxy/internal/geometry"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"strconv"
	"time"
)

type RotationSizeMode int

const (
	SizeNone RotationSizeMode = 0
	SizeFit  RotationSizeMode = 1
	SizeFill RotationSizeMode = 2
)

type RotationParams struct {
	Rotation *float64         `json:"rotation,omitempty"`
	SizeMode RotationSizeMode `json:"sizeMode,omitempty"`
}

func (*RotationParams) Params() []string {
	return []string{"rot"}
}

func (opts *RotationParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false
	if len(options) >= 1 {
		r, err := strconv.ParseFloat(options[0], 64)
		if err == nil {
			opts.Rotation = &r
		}
	}

	if len(options) >= 2 {
		switch options[1] {
		case "fit":
			opts.SizeMode = SizeFit
		case "fill":
			opts.SizeMode = SizeFill
		default:
			opts.SizeMode = SizeNone
		}
	}

	return
}

func (opts *RotationParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.Rotation == nil {
		return sourceImage, nil
	}

	if *opts.Rotation == 0 {
		return sourceImage, nil
	}

	defer utils.TrackTime(time.Now(), "Rotation")

	sw := sourceImage.Width()
	sh := sourceImage.Height()

	transparent := params.Export != nil && params.Export.Format != nil && (*params.Export.Format == "png" || *params.Export.Format == "webp")
	color, colorErr := ParseHexColor(utils.IfNil(params.Background.Color, "#00000000"))
	if colorErr != nil {
		color = ColorRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	if transparent && !sourceImage.HasAlpha() {
		err := sourceImage.AddAlpha()
		if err != nil {
			return sourceImage, err
		}
	}

	err := sourceImage.Similarity(1.0, *opts.Rotation, &vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A}, 0, 0, 0, 0)
	if err != nil {
		return sourceImage, err
	}

	if opts.SizeMode == SizeFit {
		newSize := geometry.SizeToFitSize(sourceImage.Width(), sourceImage.Height(), sw, sh)
		err := sourceImage.ThumbnailWithSize(newSize.Width, newSize.Height, vips.InterestingNone, vips.SizeBoth)
		if err != nil {
			return sourceImage, err
		}

		if sourceImage.Width() != sw || sourceImage.Height() != sh {
			if transparent {
				_ = sourceImage.EmbedBackgroundRGBA(
					(sw-sourceImage.Width())/2,
					(sh-sourceImage.Height())/2,
					sw,
					sh,
					&vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A},
				)
			} else {
				_ = sourceImage.EmbedBackground(
					(sw-sourceImage.Width())/2,
					(sh-sourceImage.Height())/2,
					sw,
					sh,
					&vips.Color{R: color.R, G: color.G, B: color.B},
				)
			}
		}
	} else if opts.SizeMode == SizeFill {
		imageRect := geometry.RectFloat64{
			X:      0,
			Y:      0,
			Width:  float64(sw),
			Height: float64(sh),
		}

		newCropSize := imageRect.SizeThatFitsRotatedRect(*opts.Rotation)
		err := sourceImage.Crop((sourceImage.Width()/2)-(newCropSize.Width/2), (sourceImage.Height()/2)-(newCropSize.Height/2), newCropSize.Width, newCropSize.Height)
		if err != nil {
			return sourceImage, err
		}
	}

	return sourceImage, nil
}
