package images

import (
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
)

func cropSmart(cW *int, cH *int, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var interesting = vips.InterestingAttention
	if params.Interesting != nil {
		interesting = *params.Interesting
	}

	err := sourceImage.SmartCrop(*cW, *cH, interesting)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}
