package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
)

func Pad(options metadata.BorderOptions, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	transparent := params.ExportParams.Format == "png" || params.ExportParams.Format == "webp"
	color, colorErr := metadata.ParseHexColor(options.Color)
	if colorErr != nil {
		color = metadata.ColorRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	ow := sourceImage.Width()
	oh := sourceImage.Height()

	newW := ow - (options.Left + options.Right)
	newH := oh - (options.Top + options.Bottom)

	newSize := geometry.SizeToFillSize(sourceImage.Width(), sourceImage.Height(), newW, newH, true)
	err := sourceImage.ThumbnailWithSize(newSize.Width, newSize.Height, vips.InterestingNone, vips.SizeDown)
	if err != nil {
		return sourceImage, err
	}

	cx := (sourceImage.Width() / 2) - (newW / 2)
	cy := (sourceImage.Height() / 2) - (newH / 2)
	err = sourceImage.Crop(cx, cy, newW, newH)
	if err != nil {
		return sourceImage, err
	}

	if transparent {
		if !sourceImage.HasAlpha() {
			err = sourceImage.AddAlpha()
			if err != nil {
				return sourceImage, err
			}
		}

		//cx := (ow / 2) - (sourceImage.Width() / 2)
		//cy := (oh / 2) - (sourceImage.Height() / 2)

		err = sourceImage.EmbedBackgroundRGBA(
			options.Left,
			options.Top,
			ow,
			oh,
			&vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A},
		)

		if err != nil {
			return sourceImage, err
		}
	} else {

		//cx := (ow / 2) - (sourceImage.Width() / 2)
		//cy := (oh / 2) - (sourceImage.Height() / 2)

		err = sourceImage.EmbedBackground(
			options.Left,
			options.Top,
			ow,
			oh,
			&vips.Color{R: color.R, G: color.G, B: color.B},
		)

		if err != nil {
			return sourceImage, err
		}
	}

	return sourceImage, nil
}
