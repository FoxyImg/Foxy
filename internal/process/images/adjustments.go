package images

import (
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"strconv"
)

type AdjustmentsParams struct {
	Order      *[]string `json:"order,omitempty"`
	Blur       *int      `json:"blur,omitempty"`
	Pixelate   *int      `json:"pixelate,omitempty"`
	Brightness *float64  `json:"brightness,omitempty"`
	Saturation *float64  `json:"saturation,omitempty"`
	Contrast   *float64  `json:"contrast,omitempty"`
	Exposure   *float64  `json:"exposure,omitempty"`
	Gamma      *float64  `json:"gamma,omitempty"`
	Hue        *float64  `json:"hue,omitempty"`
}

func (*AdjustmentsParams) Params() []string {
	return []string{"bri", "sat", "hue", "con", "exp", "gamma"}
}

func (opts *AdjustmentsParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	switch param {
	case "bri":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err == nil {
			opts.Brightness = utils.Ptr(float64(b) / 100.0)
		}
	case "sat":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err == nil {
			opts.Saturation = utils.Ptr(float64(b) / 100.0)
		}
	case "con":
		if len(options) != 1 {
			return
		}

		b, err := strconv.ParseFloat(options[0], 64)
		if err == nil {
			opts.Contrast = utils.Ptr(b)
		}
	case "exp":
		if len(options) != 1 {
			return
		}

		b, err := strconv.ParseFloat(options[0], 64)
		if err == nil {
			opts.Exposure = utils.Ptr(b)
		}
	case "gamma":
		if len(options) != 1 {
			return
		}

		b, err := strconv.ParseFloat(options[0], 64)
		if err == nil {
			opts.Gamma = utils.Ptr(b)
		}
	case "hue":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err == nil {
			opts.Hue = utils.Ptr(float64(b))
		}
	}

	return
}

func (opts *AdjustmentsParams) Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	b := utils.IfNil(opts.Brightness, 1)
	s := utils.IfNil(opts.Saturation, 1)
	h := utils.IfNil(opts.Hue, 0)

	if b != 1 || s != 1 || h != 0 {
		err := sourceImage.Modulate(b, s, h)
		if err != nil {
			return sourceImage, err
		}
	}

	c := utils.IfNil(opts.Contrast, 1)
	if c != 1 {
		err := sourceImage.Linear1(c, -0.5*c-0.5)
		if err != nil {
			return sourceImage, err
		}
	}

	exp := utils.IfNil(opts.Exposure, 0)
	if exp != 0 {
		err := sourceImage.Linear1(math.Pow(2, exp), 0)
		if err != nil {
			return sourceImage, err
		}
	}

	gamma := utils.IfNil(opts.Gamma, 1)
	log.Println("Gamma: ", gamma)
	if gamma > 0 && gamma != 1 {
		currentCS := sourceImage.ColorSpace()
		err := sourceImage.ToColorSpace(vips.InterpretationRGB16)
		if err != nil {
			return sourceImage, err
		}

		//if gamma != 0 {
		//	gamma = 1.0 / gamma
		//}

		err = sourceImage.Gamma(gamma)
		if err != nil {
			log.Println(err)
			return sourceImage, err
		}

		err = sourceImage.ToColorSpace(currentCS)
		if err != nil {
			return sourceImage, err
		}
	}

	return sourceImage, nil
}
