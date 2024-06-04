package images

import (
	"encoding/base64"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"strconv"
)

type WatermarkDropShadowParams struct {
	Opacity *float64 `json:"opacity,omitempty"`
	Blur    *float64 `json:"blur,omitempty"`
	Color   *string  `json:"color,omitempty"`
	OffsetX *int     `json:"offsetX,omitempty"`
	OffsetY *int     `json:"offsetY,omitempty"`
}

type WatermarkParams struct {
	Text       *string                    `json:"text,omitempty"`
	Font       *string                    `json:"font,omitempty"`
	VAlign     *string                    `json:"vAlign,omitempty"`
	HAlign     *string                    `json:"hAlign,omitempty"`
	Width      *float64                   `json:"width,omitempty"`
	Height     *float64                   `json:"height,omitempty"`
	Opacity    *float64                   `json:"opacity,omitempty"`
	Color      *string                    `json:"color,omitempty"`
	HPadding   *float64                   `json:"hPadding,omitempty"`
	VPadding   *float64                   `json:"vPadding,omitempty"`
	Rotate     *vips.Angle                `json:"rotate,omitempty"`
	DropShadow *WatermarkDropShadowParams `json:"dropShadow,omitempty"`
}

func (*WatermarkParams) Params() []string {
	return []string{"wm"}
}

func (opts *WatermarkDropShadowParams) ParseDropShadowParams(options []string) {
	if len(options) == 0 {
		return
	}

	switch options[0] {
	case "opacity":
		if len(options) < 2 {
			return
		}

		w, err := strconv.ParseFloat(options[1], 64)
		if err != nil {
			return
		}

		w = w / 100.0
		opts.Opacity = &w
	case "blur":
		if len(options) < 2 {
			return
		}

		w, err := strconv.ParseFloat(options[1], 64)
		if err != nil {
			return
		}

		opts.Blur = &w
	case "color":
		if len(options) < 2 {
			return
		}

		opts.Color = &options[1]
	case "offs":
		if len(options) == 2 {
			x, err := strconv.Atoi(options[1])
			if err != nil {
				return
			}

			opts.OffsetX = &x
			opts.OffsetY = &x
		} else if len(options) == 3 {
			x, err := strconv.Atoi(options[1])
			if err != nil {
				return
			}

			opts.OffsetX = &x

			y, err := strconv.Atoi(options[2])
			if err != nil {
				return
			}

			opts.OffsetY = &y
		}
	}
}

func (opts *WatermarkParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	if len(options) == 0 {
		return
	}

	if opts.HAlign == nil || *opts.HAlign == "" {
		opts.HAlign = utils.Ptr("right")
	}

	if opts.VAlign == nil || *opts.VAlign == "" {
		opts.VAlign = utils.Ptr("bottom")
	}

	switch options[0] {
	case "text":
		if len(options) < 2 {
			return
		}

		for len(options[1])%4 != 0 {
			options[1] += "="
		}

		text, err := base64.URLEncoding.DecodeString(options[1])
		if err != nil {
			return
		}

		opts.Text = utils.Ptr(string(text))
	case "font":
		if len(options) < 2 {
			return
		}

		for len(options[1])%4 != 0 {
			options[1] += "="
		}

		font, err := base64.URLEncoding.DecodeString(options[1])
		if err != nil {
			return
		}

		opts.Font = utils.Ptr(string(font))
	case "align":
		if len(options) == 2 {
			opts.HAlign = &options[1]
			opts.VAlign = &options[1]
		} else if len(options) == 3 {
			opts.HAlign = &options[1]
			opts.VAlign = &options[2]
		}
	case "size":
		if len(options) == 2 {
			w, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			opts.Width = &w
			opts.Height = &w
		} else if len(options) == 3 {
			w, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			opts.Width = &w

			h, err := strconv.ParseFloat(options[2], 64)
			if err != nil {
				return
			}

			opts.Height = &h
		}
	case "opacity":
		if len(options) == 2 {
			w, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			w = w / 100.0
			opts.Opacity = &w
		}
	case "color":
		if len(options) == 2 {
			opts.Color = &options[1]
		}
	case "rot":
		if len(options) < 2 {
			return
		}

		switch options[1] {
		case "90":
			opts.Rotate = utils.Ptr(vips.Angle90)
		case "180":
			opts.Rotate = utils.Ptr(vips.Angle180)
		case "270":
			opts.Rotate = utils.Ptr(vips.Angle270)
		}
	case "pad":
		if len(options) == 2 {
			p, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			opts.HPadding = &p
			opts.VPadding = &p
		} else if len(options) == 3 {
			h, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			opts.HPadding = &h

			v, err := strconv.ParseFloat(options[2], 64)
			if err != nil {
				return
			}

			opts.VPadding = &v
		}
	case "shadow":
		if opts.DropShadow == nil {
			opts.DropShadow = &WatermarkDropShadowParams{}
		}

		opts.DropShadow.ParseDropShadowParams(options[1:])
	}

	return
}

func (opts *WatermarkParams) CreateWatermarkFillImage(offx int, offy int, color string, blur float64, opacity float64, watermarkTextImage *vips.ImageRef) (*vips.ImageRef, error) {
	watermarkTextImageCopy, err := watermarkTextImage.Copy()
	if err != nil {
		return nil, err
	}

	watermarkColor, err := vips.Black(watermarkTextImage.Width(), watermarkTextImage.Height())
	if err != nil {
		return nil, err
	}

	_ = watermarkColor.Cast(vips.BandFormatUchar)

	parsedColor, err := ParseHexColor(color)
	if err != nil {
		return nil, err
	}

	_ = watermarkColor.BandJoinConst([]float64{0, 0})
	err = watermarkColor.ToColorSpace(vips.InterpretationSRGB)
	if err != nil {
		log.Println("watermark Color Space Error:", err)
	}

	_ = watermarkColor.DrawRect(vips.ColorRGBA{R: parsedColor.R, G: parsedColor.G, B: parsedColor.B, A: 255}, 0, 0, watermarkTextImage.Width(), watermarkTextImage.Height(), true)
	_ = watermarkTextImageCopy.ExtractBand(0, 1)
	_ = watermarkColor.BandJoin(watermarkTextImageCopy)
	if watermarkColor.Bands() > 4 {
		_ = watermarkColor.ExtractBand(0, 4)
	}

	//TODO: Fix this.  For some reason if I don't do this conversion I get an error "no route from multiband to srgb".  Googling is unhelpful.
	buffer, _, err := watermarkColor.ExportPng(nil)
	if err != nil {
		return watermarkColor, err
	}

	multibandWTF, err := vips.NewImageFromBuffer(buffer)
	if err != nil {
		return watermarkColor, err
	}

	watermarkColor = multibandWTF

	if blur > 0 {
		_ = watermarkColor.GaussianBlur(blur)
	}

	if opacity < 1.0 {
		_ = watermarkColor.Linear([]float64{1.0, 1.0, 1.0, opacity}, []float64{0.0, 0.0, 0.0, 0.0})
	}

	if offx > 0 || offy > 0 {
		ow := watermarkTextImage.Width()
		oh := watermarkTextImage.Height()

		_ = watermarkColor.Crop(0, 0, ow-offx, oh-offy)
		_ = watermarkColor.EmbedBackgroundRGBA(offx, offy, ow, oh, &vips.ColorRGBA{R: 0, G: 0, B: 0, A: 0})
	}

	return watermarkColor, nil
}

func (opts *WatermarkParams) Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	log.Println("format?", sourceImage.BandFormat())

	if opts.Text == nil || *opts.Text == "" || opts.Font == nil || opts.Width == nil || *opts.Width == 0 || opts.Height == nil || *opts.Height == 0 {
		return sourceImage, nil
	}

	watermarkText, err := vips.Black(sourceImage.Width(), sourceImage.Height())
	if err != nil {
		return sourceImage, err
	}

	vpadding := utils.Max(4, int(utils.IfNil(opts.VPadding, 10.0)))
	hpadding := utils.Max(4, int(utils.IfNil(opts.HPadding, 10.0)))

	textWidth := (*opts.Width / 100.0) * float64(sourceImage.Width()-(hpadding*2))
	textHeight := (*opts.Height / 100.0) * float64(sourceImage.Height()-(vpadding*2))
	if opts.Rotate != nil && (*opts.Rotate == vips.Angle90 || *opts.Rotate == vips.Angle270) {
		textWidth = (*opts.Width / 100.0) * float64(sourceImage.Height()-(hpadding*2))
		textHeight = (*opts.Height / 100.0) * float64(sourceImage.Width()-(vpadding*2))
	}

	err = watermarkText.Label(&vips.LabelParams{
		Text:      *opts.Text,
		Font:      *opts.Font,
		OffsetX:   vips.Scalar{Value: 0, Relative: false},
		OffsetY:   vips.Scalar{Value: 0, Relative: false},
		Width:     vips.Scalar{Value: textWidth, Relative: false},
		Height:    vips.Scalar{Value: textHeight, Relative: false},
		Color:     vips.Color{R: 255, G: 255, B: 255},
		Alignment: vips.AlignCenter,
		Opacity:   1,
	})
	if err != nil {
		log.Println("Label Error:", err)
		return sourceImage, err
	}

	left, top, width, height, err := watermarkText.FindTrim(0, &vips.Color{R: 0, G: 0, B: 0})
	if err != nil {
		log.Println("Find Trim Error:", err)
		return sourceImage, err
	}

	_ = watermarkText.Crop(left, top, width, height)
	_ = watermarkText.EmbedBackground(hpadding, vpadding, watermarkText.Width()+(hpadding*2), watermarkText.Height()+(vpadding*2), &vips.Color{R: 0, G: 0, B: 0})
	_ = watermarkText.ExtractBand(0, watermarkText.Bands()-1)

	watermarkColor, err := opts.CreateWatermarkFillImage(0, 0, utils.IfNil(opts.Color, "FFFFFF"), 0, 1, watermarkText)
	if err != nil {
		return sourceImage, err
	}

	if opts.DropShadow != nil {
		watermarkShadow, err := opts.CreateWatermarkFillImage(
			utils.IfNil(opts.DropShadow.OffsetX, 0),
			utils.IfNil(opts.DropShadow.OffsetY, 0),
			utils.IfNil(opts.DropShadow.Color, "000000"),
			utils.IfNil(opts.DropShadow.Blur, 3),
			utils.IfNil(opts.DropShadow.Opacity, 1.0),
			watermarkText,
		)
		if err != nil {
			return sourceImage, err
		}

		_ = watermarkShadow.Composite(watermarkColor, vips.BlendModeOver, 0, 0)

		if watermarkShadow.Bands() > 4 {
			_ = watermarkShadow.ExtractBand(0, 4)
		}

		watermarkColor = watermarkShadow
	}

	if opts.Rotate != nil && *opts.Rotate != vips.Angle0 {
		_ = watermarkColor.Rotate(*opts.Rotate)
	}

	x := 0
	y := 0
	if opts.HAlign != nil && *opts.HAlign == "center" {
		x = (sourceImage.Width() - watermarkColor.Width()) / 2
	} else if opts.HAlign != nil && *opts.HAlign == "right" {
		x = sourceImage.Width() - watermarkColor.Width()
	}

	if opts.VAlign != nil && *opts.VAlign == "center" {
		y = (sourceImage.Height() - watermarkColor.Height()) / 2
	} else if opts.VAlign != nil && *opts.VAlign == "bottom" {
		y = sourceImage.Height() - watermarkColor.Height()
	}

	if !sourceImage.HasAlpha() {
		_ = sourceImage.AddAlpha()
	}

	if opts.Opacity != nil && *opts.Opacity > 0 {
		_ = watermarkColor.Linear([]float64{1.0, 1.0, 1.0, *opts.Opacity}, []float64{0.0, 0.0, 0.0, 0.0})
	}

	err = sourceImage.Composite(watermarkColor, vips.BlendModeOver, x, y)
	if err != nil {
		log.Println("Composite Error:", err)
	}

	return sourceImage, nil

}
