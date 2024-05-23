package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
)

func cropPerson(cW *int, cH *int, imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var faceBounds geometry.Box

	var personLabels = metadata.GetPersonLabels(imageMeta.Labels)
	if params.PersonIndex != nil && *params.PersonIndex < len(personLabels) {
		faceBounds = *personLabels[*params.PersonIndex].Box
	} else {
		faceBounds = metadata.CalcPersonBounds(imageMeta.Labels)
	}

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

	cx := int((faceBounds.Left + (faceBounds.Width / 2.0)) * float64(sw))
	cy := int((faceBounds.Top + (faceBounds.Height / 2.0)) * float64(sh))

	cropX := utils.Min(sw, utils.Max(0, cx-(cropSize.Width/2)))
	cropY := utils.Min(sh, utils.Max(0, cy-(cropSize.Height/2)))

	if (cropX + cropSize.Width) > sw {
		cropX = utils.Max(0, sw-cropSize.Width)
	}

	if (cropY + cropSize.Height) > sh {
		cropY = utils.Max(0, sh-cropSize.Height)
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

	if (cropSize.Width > targetWidth) || (cropSize.Height > targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}
