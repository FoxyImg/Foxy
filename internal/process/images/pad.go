package images

import (
	"foxy/internal/config"
	"foxy/internal/geometry"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
)

type PadOptions struct {
	BorderOptions
}

func (*PadOptions) Params() []string {
	return []string{"pad"}
}

func (opt *PadOptions) Process(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	l := utils.IfNil(opt.Left, 0)
	t := utils.IfNil(opt.Top, 0)
	r := utils.IfNil(opt.Right, 0)
	b := utils.IfNil(opt.Bottom, 0)

	if l == 0 && t == 0 && r == 0 && b == 0 {
		return sourceImage, nil
	}

	transparent := params.Export != nil && params.Export.Format != nil && (*params.Export.Format == "png" || *params.Export.Format == "webp")
	color, colorErr := ParseHexColor(utils.IfNil(opt.Color, "#00000000"))
	if colorErr != nil {
		color = ColorRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	ow := sourceImage.Width()
	oh := sourceImage.Height()

	newW := ow - (l + r)
	newH := oh - (t + b)

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

		err = sourceImage.EmbedBackgroundRGBA(
			l,
			t,
			ow,
			oh,
			&vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A},
		)

		if err != nil {
			return sourceImage, err
		}
	} else {
		err = sourceImage.EmbedBackground(
			l,
			t,
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
