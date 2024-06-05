package images

import (
	"foxy/internal/config"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"strconv"
)

type BorderOptions struct {
	Color  *string `json:"color,omitempty"`
	Left   *int    `json:"left,omitempty"`
	Top    *int    `json:"top,omitempty"`
	Right  *int    `json:"right,omitempty"`
	Bottom *int    `json:"bottom,omitempty"`
}

func (*BorderOptions) Params() []string {
	return []string{"border"}
}

func (opt *BorderOptions) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false
	if len(options) <= 1 || len(options) == 4 {
		return
	}

	opt.Color = utils.Ptr("#" + options[0])

	if len(options) == 2 {
		p, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		opt.Left = utils.Ptr(p)
		opt.Top = utils.Ptr(p)
		opt.Right = utils.Ptr(p)
		opt.Bottom = utils.Ptr(p)
	} else if len(options) == 3 {
		h, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		v, err := strconv.Atoi(options[2])
		if err != nil {
			return
		}

		opt.Left = utils.Ptr(h)
		opt.Top = utils.Ptr(v)
		opt.Right = utils.Ptr(h)
		opt.Bottom = utils.Ptr(v)
	} else if len(options) == 5 {
		l, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		t, err := strconv.Atoi(options[2])
		if err != nil {
			return
		}

		r, err := strconv.Atoi(options[3])
		if err != nil {
			return
		}

		b, err := strconv.Atoi(options[4])
		if err != nil {
			return
		}

		opt.Left = utils.Ptr(l)
		opt.Top = utils.Ptr(t)
		opt.Right = utils.Ptr(r)
		opt.Bottom = utils.Ptr(b)
	}

	return
}

func (opt *BorderOptions) Process(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
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

	if transparent {
		if !sourceImage.HasAlpha() {
			err := sourceImage.AddAlpha()
			if err != nil {
				return sourceImage, err
			}
		}

		err := sourceImage.EmbedBackgroundRGBA(
			l,
			t,
			sourceImage.Width()+l+r,
			sourceImage.Height()+t+b,
			&vips.ColorRGBA{R: color.R, G: color.G, B: color.B, A: color.A},
		)

		if err != nil {
			return sourceImage, err
		}
	} else {
		err := sourceImage.EmbedBackground(
			l,
			t,
			sourceImage.Width()+l+r,
			sourceImage.Height()+t+b,
			&vips.Color{R: color.R, G: color.G, B: color.B},
		)

		if err != nil {
			return sourceImage, err
		}
	}

	return sourceImage, nil
}
