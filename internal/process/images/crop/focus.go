package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"log"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func cropFocus(focusX float64, focusY float64, cW *int, cH *int, params *metadata.ImageParams, sourceImage *vips.ImageRef, zoom *float64) (*vips.ImageRef, error) {
	if params == nil {
		return cropFill(cW, cH, params, sourceImage)
	}

	sw := sourceImage.Width()
	sh := sourceImage.Height()

	fpx := int(math.Round(focusX * float64(sw)))
	fpy := int(math.Round(focusY * float64(sh)))

	var newCW int
	var newCH int
	if *cW < *cH {
		newCH = utils.Max(*cH, utils.Min(fpy, sw-fpy))
		newCW = int(math.Round(float64(newCH) * (float64(*cW) / float64((*cH)))))
	} else if *cW > *cH {
		newCW = utils.Max(*cW, utils.Min(fpx, sw-fpx))
		newCH = int(math.Round(float64(newCW) * (float64(*cH) / float64(*cW))))
	} else {
		newCW = utils.Max(*cW, utils.Min(fpx, sw-fpx, fpy, sh-fpy))
		newCH = newCW
	}

	if zoom != nil && *zoom > 0 {
		newCW = int(math.Round(float64(newCW) * (1.0 / *zoom)))
		newCH = int(math.Round(float64(newCH) * (1.0 / *zoom)))
	}

	var cropSize geometry.Size
	if newCW*2 > sw || newCH*2 > sh {
		cropSize = geometry.SizeToFitSize(newCW*2, newCH*2, sw, sh)
	} else {
		cropSize = geometry.Size{Width: newCW * 2, Height: newCH * 2}
	}

	if params.Zoom != nil {
		cropSize.Width = int(math.Round(float64(cropSize.Width) * (1 / *params.Zoom)))
		cropSize.Height = int(math.Round(float64(cropSize.Height) * (1 / *params.Zoom)))
	}

	cropX := utils.Min(sw-cropSize.Width, utils.Max(0, fpx-(cropSize.Width/2)))
	cropY := utils.Min(sh-cropSize.Height, utils.Max(0, fpy-(cropSize.Height/2)))

	if params.HGravity == "left" {
		cropX = 0
	} else if params.HGravity == "right" {
		cropX = sw - cropSize.Width
	}

	if params.VGravity == "top" {
		cropY = 0
	} else if params.VGravity == "bottom" {
		cropY = sh - cropSize.Height
	}

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width != *cW) || (cropSize.Height != *cH) {
		err = sourceImage.ThumbnailWithSize(*cW, *cH, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}
