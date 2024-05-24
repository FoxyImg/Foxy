package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
)

func cropFill(cW *int, cH *int, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var targetWidth int
	var targetHeight int

	if cW == nil {
		targetWidth = 10000
	} else {
		targetWidth = *cW
	}

	if cH == nil {
		targetHeight = 10000
	} else {
		targetHeight = *cH
	}

	cropSize := geometry.SizeToFitSize(targetWidth, targetHeight, sw, sh)
	if params.Zoom != nil {
		cropSize.Width = int(math.Floor(float64(cropSize.Width) * (1 / *params.Zoom)))
		cropSize.Height = int(math.Floor(float64(cropSize.Height) * (1 / *params.Zoom)))
	}

	cropX := utils.Min(sw, utils.Max(0, (sw/2)-(cropSize.Width/2)))
	cropY := utils.Min(sh, utils.Max(0, (sh/2)-(cropSize.Height/2)))

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width != targetWidth) || (cropSize.Height != targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}
