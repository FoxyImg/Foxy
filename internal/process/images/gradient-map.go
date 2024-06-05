package images

import (
	"fmt"
	"foxy/internal/config"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"strconv"
	"strings"
)

type GradientStop struct {
	Stop  *float64 `json:"stop,omitempty"`
	Color *string  `json:"color,omitempty"`
}

type GradientMapParams struct {
	Opacity    *float64        `json:"opacity,omitempty"`
	BlendMode  *vips.BlendMode `json:"blendMode,omitempty"`
	Stops      *[]GradientStop `json:"stops,omitempty"`
	Blur       *float64        `json:"blur,omitempty"`
	Monochrome *bool           `json:"monochrome,omitempty"`
}

func (*GradientMapParams) Params() []string {
	return []string{"gm"}
}

func (opts *GradientMapParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	switch param {
	case "gm":
		if options[0] == "mono" && len(options) == 2 {
			opts.Monochrome = utils.Ptr(options[1] == "true")
		} else if options[0] == "opacity" && len(options) == 2 {
			opacity, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			opts.Opacity = utils.Ptr(opacity / 100.0)
		} else if options[0] == "blend" && len(options) == 2 {
			mode, err := strconv.Atoi(options[1])
			if err != nil {
				return
			}

			opts.BlendMode = utils.Ptr(vips.BlendMode(mode))
		} else if options[0] == "blur" && len(options) == 2 {
			blur, err := strconv.Atoi(options[1])
			if err != nil {
				return
			}

			opts.Blur = utils.Ptr(float64(blur))
		} else if options[0] == "stops" && len(options) > 2 {
			stops := []GradientStop{}
			for _, stopDef := range options[1:] {
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

			opts.Stops = &stops
		}
	}

	return
}

func (opts *GradientMapParams) Process(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.Opacity != nil && *opts.Opacity > 0 && opts.Stops != nil && len(*opts.Stops) > 1 {
		bm := utils.IfNil(opts.BlendMode, vips.BlendModeOver)

		if !sourceImage.HasAlpha() {
			_ = sourceImage.AddAlpha()
		}

		stepsSvg := ""
		for _, stop := range *opts.Stops {
			stepsSvg += fmt.Sprintf("<stop offset='%d%%' stop-color='%s'/>", int(*stop.Stop), "#"+*stop.Color)
		}

		mapSvg := fmt.Sprintf("<svg width='256' height='1' xmlns='http://www.w3.org/2000/svg'><defs><linearGradient id='grad' x1='0%%' y1='0%%' x2='100%%' y2='0%%'>%s</linearGradient></defs><rect width='256' height='1' fill='url(#grad)'/></svg>", stepsSvg)
		mapImg, mapImgErr := vips.LoadImageFromBuffer([]byte(mapSvg), vips.NewImportParams())
		if mapImgErr != nil {
			return sourceImage, mapImgErr
		}

		mappedImg, _ := sourceImage.Copy()
		if utils.IfNil(opts.Monochrome, true) {
			_ = mappedImg.Modulate(1, 0, 0)
		}
		err := mappedImg.Maplut(mapImg)
		blur := utils.IfNil(opts.Blur, 0)
		if blur > 0 {
			_ = mappedImg.GaussianBlur(blur)
		}

		if err != nil {
			return sourceImage, err
		}

		_ = mappedImg.Linear([]float64{1.0, 1.0, 1.0, 0}, []float64{0.0, 0.0, 0.0, 255.0 * *opts.Opacity})
		_ = sourceImage.Composite(mappedImg, bm, 0, 0)
	}

	return sourceImage, nil
}
