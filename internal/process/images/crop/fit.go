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
		backgroundColor, bgColorErr := metadata.ParseHexColor(params.BackgroundColor)
		if bgColorErr != nil {
			backgroundColor = metadata.ColorRGBA{R: 0, G: 0, B: 0, A: 0}
		}

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
				&vips.ColorRGBA{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B, A: backgroundColor.A},
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
				&vips.Color{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B},
			)

			if err != nil {
				return sourceImage, err
			}
		}
	}

	return sourceImage, nil
}
