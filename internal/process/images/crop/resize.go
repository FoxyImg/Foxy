package images

import (
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
)

func resize(params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var targetWidth int
	var targetHeight int

	if params.Width == nil {
		targetWidth = 10000
	} else {
		targetWidth = *params.Width
	}

	if params.Height == nil {
		targetHeight = 10000
	} else {
		targetHeight = *params.Height
	}

	err := sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeBoth)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}
