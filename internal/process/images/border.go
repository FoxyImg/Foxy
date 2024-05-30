package images

import (
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
)

func Border(options metadata.BorderOptions, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	transparent := params.ExportParams.Format == "png" || params.ExportParams.Format == "webp"
	color, colorErr := metadata.ParseHexColor(options.Color)
	if colorErr != nil {
		color = metadata.ColorRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	if transparent {
		if !sourceImage.HasAlpha() {
			err := sourceImage.AddAlpha()
			if err != nil {
				return sourceImage, err
			}
		}

		err := sourceImage.EmbedBackgroundRGBA(
			options.Left,
			options.Top,
			sourceImage.Width()+options.Left+options.Right,
			sourceImage.Height()+options.Top+options.Bottom,
			&vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A},
		)

		if err != nil {
			return sourceImage, err
		}
	} else {
		err := sourceImage.EmbedBackground(
			options.Left,
			options.Top,
			sourceImage.Width()+options.Left+options.Right,
			sourceImage.Height()+options.Top+options.Bottom,
			&vips.Color{R: color.R, G: color.G, B: color.B},
		)

		if err != nil {
			return sourceImage, err
		}
	}

	return sourceImage, nil
}
