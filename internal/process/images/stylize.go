package images

import (
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"slices"
	"strconv"
	"strings"
)

type StylizeParams struct {
	Order    *[]string `json:"order,omitempty"`
	Blur     *int      `json:"blur,omitempty"`
	Pixelate *int      `json:"pixelate,omitempty"`
}

func (*StylizeParams) Params() []string {
	return []string{"stylize", "blur", "px"}
}

func (opts *StylizeParams) ParseParams(param string, options []string) (needsVision bool) {
	if opts.Order == nil {
		opts.Order = &[]string{"blur", "px"}
	}

	needsVision = false

	switch param {
	case "blur":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		opts.Blur = &b
	case "px":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		opts.Pixelate = &b
	case "stylize":
		if len(options) != 2 {
			return
		}

		if options[0] == "order" {
			opts.Order = utils.Ptr(strings.Split(options[1], ","))
			if !slices.Contains(*opts.Order, "blur") {
				*opts.Order = append(*opts.Order, "blur")
			}
			if !slices.Contains(*opts.Order, "px") {
				*opts.Order = append(*opts.Order, "px")
			}
		}
	}

	return
}

func (opts *StylizeParams) Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	blur := utils.IfNil(opts.Blur, 0)
	pixelate := utils.IfNil(opts.Pixelate, 0)

	if opts.Order == nil || len(*opts.Order) == 0 || (blur == 0 && pixelate == 0) {
		return sourceImage, nil
	}

	for _, order := range *opts.Order {
		if order == "blur" && blur > 0 {
			err := sourceImage.GaussianBlur(float64(blur))
			if err != nil {
				return sourceImage, err
			}
		} else if order == "px" && pixelate > 0 {
			_ = sourceImage.Resize(1.0/float64(pixelate), vips.KernelLanczos3)
			_ = sourceImage.Resize(float64(pixelate), vips.KernelNearest)
		}
	}

	return sourceImage, nil
}
