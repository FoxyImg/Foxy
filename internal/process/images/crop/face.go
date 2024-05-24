package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
)

func cropFace(cW *int, cH *int, imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var faceBounds geometry.Box
	if params.Face.Index != nil && *params.Face.Index < len(imageMeta.Faces) {
		faceBounds = imageMeta.Faces[*params.Face.Index].Box
	} else {
		if params.Face.Largest != nil {
			area := 0.0
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea > area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else if params.Face.Smallest != nil {
			area := math.MaxFloat64
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea < area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else {
			faceBounds = metadata.CalcFacesBounds(imageMeta.Faces)
		}
	}

	fy := int(math.Floor(faceBounds.Top * float64(sh)))
	fh := int(math.Floor(faceBounds.Height * float64(sh)))

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

	cx := int((faceBounds.Left + (faceBounds.Width / 2.0)) * float64(sw))
	var cy int
	if fh > cropSize.Height {
		cy = fy + int(math.Floor(float64(fh)/2.0))
	} else {
		cy = fy - int(math.Floor((float64(params.Face.Padding)/512.0)*float64(sh)))
	}

	cropX := utils.Min(sw, utils.Max(0, cx-(cropSize.Width/2)))
	var cropY int
	if fh > cropSize.Height {
		cropY = utils.Min(sw, utils.Max(0, cy-(cropSize.Height/2)))
	} else {
		cropY = utils.Min(sh, utils.Max(0, cy))
	}

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

	if (cropSize.Width != targetWidth) || (cropSize.Height != targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeBoth)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}
