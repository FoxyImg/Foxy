package params

import (
	"fmt"
	"foxy/internal/config"
	"foxy/internal/storage"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
)

type MaskParams struct {
	Type         *string `json:"type,omitempty"`
	CornerRadius *int    `json:"cornerRadius,omitempty"`
	ImageKey     *string `json:"url,omitempty"`
	Sizing       *string `json:"sizing,omitempty"`
}

func (*MaskParams) Params() []string {
	return []string{"mask"}
}

func (opts *MaskParams) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false

	//mask:rect:12
	//mask:circle
	//mask:image:<url>:fit
	switch options[0] {
	case "rect":
		fallthrough
	case "square":
		if len(options) == 2 {
			opts.Type = &options[0]
			opts.CornerRadius = IntVal(options[1:])
		}
	case "ellipse":
		fallthrough
	case "circle":
		opts.Type = &options[0]
	case "image":
		if len(options) == 3 {
			opts.Type = utils.Ptr("image")
			opts.ImageKey = DecodeBase64StringVal(options[1:])
			opts.Sizing = utils.Ptr(options[2])
		}
	}

	return
}

func (opts *MaskParams) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if opts.Type == nil {
		return sourceImage, nil
	}

	var maskImg *vips.ImageRef
	if *opts.Type != "image" {
		rx := utils.Max(sourceImage.Width(), sourceImage.Height())
		if opts.CornerRadius != nil {
			rx = *opts.CornerRadius
		}

		dx := 0
		dy := 0
		dw := sourceImage.Width()
		dh := sourceImage.Height()

		if *opts.Type == "circle" || *opts.Type == "square" {
			md := utils.Min(dw, dh)
			dx = (dw - md) / 2
			dy = (dh - md) / 2
			dw = md
			dh = md
		}

		maskRect := fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"%d\" style=\"fill:black;\" />",
			dx,
			dy,
			dw,
			dh,
			rx,
		)

		maskSVGString := fmt.Sprintf("<svg width=\"%d\" height=\"%d\" viewport=\"0 0 %d %d\" xmlns=\"http://www.w3.org/2000/svg\">%s</svg>",
			sourceImage.Width(),
			sourceImage.Height(),
			sourceImage.Width(),
			sourceImage.Height(),
			maskRect,
		)

		svg, err := utils.RenderSVG(maskSVGString) // vips.LoadImageFromBuffer([]byte(svgString), vips.NewImportParams())
		if err != nil {
			return sourceImage, err
		}

		log.Println("Mask SVG:", maskSVGString)
		_ = DumpDebugImage("svg-mask", svg)

		maskImg, err = svg.ExtractBandToImage(3, 1)
	} else {
		oimg, err := storage.GetSourceImage(config, sourceId, *opts.ImageKey, params.Debug.DisableSourceCache)
		if err != nil {
			return sourceImage, err
		}

		if oimg.HasAlpha() {
			maskImg, err = oimg.ExtractBandToImage(3, 1)
		} else {
			maskImg, err = oimg.ExtractBandToImage(0, 1)
		}

		hs := float64(sourceImage.Width()) / float64(maskImg.Width())
		vs := float64(sourceImage.Height()) / float64(maskImg.Height())
		if *opts.Sizing == "fit" {
			s := math.Min(hs, vs)
			_ = maskImg.Resize(s, vips.KernelLanczos3)
			_ = maskImg.EmbedBackgroundRGBA(
				(sourceImage.Width()/2)-(maskImg.Width()/2),
				(sourceImage.Height()/2)-(maskImg.Height()/2),
				sourceImage.Width(),
				sourceImage.Height(),
				&vips.ColorRGBA{R: 0, G: 0, B: 0, A: 0},
			)

		} else if *opts.Sizing == "fill" {
			s := math.Max(hs, vs)
			_ = maskImg.Resize(s, vips.KernelLanczos3)
			_ = maskImg.Crop(
				(maskImg.Width()/2)-(sourceImage.Width()/2),
				(maskImg.Height()/2)-(sourceImage.Height()/2),
				sourceImage.Width(),
				sourceImage.Height(),
			)
		} else if *opts.Sizing == "stretch" {
			if hs != 1.0 || vs != 1.0 {
				_ = maskImg.ResizeWithVScale(hs, vs, vips.KernelLanczos3)
			}
		}
	}

	if sourceImage.HasAlpha() {
		_ = sourceImage.ExtractBand(0, sourceImage.Bands()-1)
	}

	_ = sourceImage.BandJoin(maskImg)

	return sourceImage, nil
}
