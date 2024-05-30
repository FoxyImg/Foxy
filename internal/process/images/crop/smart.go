package images

import (
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
	"math"
)

func cropSmart(cW *int, cH *int, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var interesting = vips.InterestingAttention
	if params.Interesting != nil {
		interesting = *params.Interesting
	}

	if params.Zoom != nil {
		*cW = int(math.Round(float64(*cW) * (1 / *params.Zoom)))
		*cH = int(math.Round(float64(*cH) * (1 / *params.Zoom)))
	}

	err := sourceImage.SmartCrop(*cW, *cH, interesting)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}
