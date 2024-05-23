package images

import (
	"foxy/internal/metadata"

	"github.com/davidbyttow/govips/v2/vips"
)

func cropFit(cW *int, cH *int, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sourceImage, err := resize(params, sourceImage)
	if err != nil {
		return sourceImage, err
	}

	if cW != nil && cH != nil {
		transparent := params.ExportParams.Format == "png" || params.ExportParams.Format == "webp"
		if transparent {
			if !sourceImage.HasAlpha() {
				err = sourceImage.AddAlpha()
				if err != nil {
					return sourceImage, err
				}
			}

			err = sourceImage.EmbedBackgroundRGBA(
				(*cW/2)-(sourceImage.Width()/2),
				(*cH/2)-(sourceImage.Height()/2),
				*cW,
				*cH,
				&vips.ColorRGBA{R: params.BackgroundColor.R, G: params.BackgroundColor.G, B: params.BackgroundColor.B, A: params.BackgroundColor.A},
			)

			if err != nil {
				return sourceImage, err
			}
		} else {
			err = sourceImage.EmbedBackground(
				(*cW/2)-(sourceImage.Width()/2),
				(*cH/2)-(sourceImage.Height()/2),
				*cW,
				*cH,
				&vips.Color{R: params.BackgroundColor.R, G: params.BackgroundColor.G, B: params.BackgroundColor.B},
			)

			if err != nil {
				return sourceImage, err
			}
		}
	}

	return sourceImage, nil
}
