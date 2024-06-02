package images

import (
	"fmt"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"slices"
	"strconv"
	"strings"
)

type GradientStop struct {
	Stop  *float64 `json:"stop,omitempty"`
	Color *string  `json:"color,omitempty"`
}

type GradientMap struct {
	Opacity   *float64        `json:"opacity,omitempty"`
	BlendMode *vips.BlendMode `json:"blendMode,omitempty"`
	Stops     *[]GradientStop `json:"stops,omitempty"`
}

type StylizeParams struct {
	Order       *[]string    `json:"order,omitempty"`
	Blur        *int         `json:"blur,omitempty"`
	Pixelate    *int         `json:"pixelate,omitempty"`
	Brightness  *float64     `json:"brightness,omitempty"`
	Saturation  *float64     `json:"saturation,omitempty"`
	Hue         *float64     `json:"hue,omitempty"`
	GradientMap *GradientMap `json:"gradientMap,omitempty"`
}

func (*StylizeParams) Params() []string {
	return []string{"stylize", "blur", "px", "bri", "sat", "hue", "gm"}
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
		if err == nil {
			opts.Blur = &b
		}
	case "px":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err == nil {
			opts.Pixelate = &b
		}

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
	case "hue":
		if len(options) != 1 {
			return
		}

		b, err := strconv.Atoi(options[0])
		if err == nil {
			opts.Hue = utils.Ptr(float64(b))
		}
	case "gm":
		if len(options) < 4 {
			return
		}

		opacity, err := strconv.ParseFloat(options[0], 64)
		if err != nil || opacity == 0 {
			return
		}

		mode, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		stops := []GradientStop{}
		for _, stopDef := range options[2:] {
			stopParts := strings.Split(stopDef, ",")
			if len(stopParts) != 2 {
				return
			}

			position, err := strconv.ParseFloat(stopParts[0], 64)
			if err != nil {
				return
			}

			stops = append(stops, GradientStop{
				Stop:  utils.Ptr(position),
				Color: utils.Ptr(stopParts[1]),
			})
		}

		if len(stops) < 2 {
			return
		}

		opts.GradientMap = &GradientMap{
			Opacity:   utils.Ptr(opacity / 100.0),
			BlendMode: utils.Ptr(vips.BlendMode(mode)),
			Stops:     &stops,
		}
	}

	return
}

func (opts *StylizeParams) Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	blur := utils.IfNil(opts.Blur, 0)
	pixelate := utils.IfNil(opts.Pixelate, 0)

	if opts.Order != nil && len(*opts.Order) != 0 && (blur != 0 || pixelate != 0) {
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
	}

	b := utils.IfNil(opts.Brightness, 1)
	s := utils.IfNil(opts.Saturation, 1)
	h := utils.IfNil(opts.Hue, 0)

	if b != 1 || s != 1 || h != 0 {
		err := sourceImage.Modulate(b, s, h)
		if err != nil {
			return sourceImage, err
		}
	}

	if opts.GradientMap != nil && opts.GradientMap.Opacity != nil && *opts.GradientMap.Opacity > 0 && opts.GradientMap.Stops != nil && len(*opts.GradientMap.Stops) > 1 {
		bm := utils.IfNil(opts.GradientMap.BlendMode, vips.BlendModeOver)

		if !sourceImage.HasAlpha() {
			_ = sourceImage.AddAlpha()
		}

		stepsSvg := ""
		for _, stop := range *opts.GradientMap.Stops {
			stepsSvg += fmt.Sprintf("<stop offset='%d%%' stop-color='%s'/>", int(*stop.Stop), "#"+*stop.Color)
		}

		mapSvg := fmt.Sprintf("<svg width='256' height='1' xmlns='http://www.w3.org/2000/svg'><defs><linearGradient id='grad' x1='0%%' y1='0%%' x2='100%%' y2='0%%'>%s</linearGradient></defs><rect width='256' height='1' fill='url(#grad)'/></svg>", stepsSvg)
		mapImg, mapImgErr := vips.LoadImageFromBuffer([]byte(mapSvg), vips.NewImportParams())
		if mapImgErr != nil {
			return sourceImage, mapImgErr
		}

		mappedImg, _ := sourceImage.Copy()
		err := mappedImg.Maplut(mapImg)
		if err != nil {
			return sourceImage, err
		}

		_ = mappedImg.Linear([]float64{1.0, 1.0, 1.0, *opts.GradientMap.Opacity}, []float64{0.0, 0.0, 0.0, 0.0})
		_ = sourceImage.Composite(mappedImg, bm, 0, 0)
	}

	return sourceImage, nil
}
