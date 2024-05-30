package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"log"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func cropBounds(cW *int, cH *int, bounds geometry.Box, cropZoom *float64, boundsZoom *float64, padding int, hGravity string, vGravity string, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sourceWidth := sourceImage.Width()
	sourceHeight := sourceImage.Height()

	if bounds.Width*bounds.Height == 0 {
		return cropFill(cW, cH, params, sourceImage)
	}

	boundsX := int(math.Round(bounds.Left * float64(sourceWidth)))
	boundsY := int(math.Round(bounds.Top * float64(sourceHeight)))
	boundsWidth := int(math.Round(bounds.Width * float64(sourceWidth)))
	boundsHeight := int(math.Round(bounds.Height * float64(sourceHeight)))

	var targetCropWidth int
	var targetCropHeight int

	if cW == nil {
		targetCropWidth = 10000
	} else {
		targetCropWidth = *cW
	}

	if cH == nil {
		targetCropHeight = 10000
	} else {
		targetCropHeight = *cH
	}

	cropSize := geometry.SizeToFitSize(targetCropWidth, targetCropHeight, sourceWidth, sourceHeight)

	actualPadding := int(math.Round((float64(padding) / 512.0) * float64(sourceWidth)))
	var zoom *float64 = nil
	if boundsZoom != nil {
		z := math.Min(float64(cropSize.Height)/float64(boundsHeight+(actualPadding*2)), float64(cropSize.Width)/float64(boundsWidth+(actualPadding*2))) * *boundsZoom
		zoom = &z
	} else if cropZoom != nil {
		zoom = cropZoom
	}

	if zoom != nil {
		newCW := int(math.Round(float64(cropSize.Width) * (1 / *zoom)))
		newCH := int(math.Round(float64(cropSize.Height) * (1 / *zoom)))
		if newCW <= sourceWidth && newCH <= sourceHeight {
			cropSize.Width = newCW
			cropSize.Height = newCH
		}
	}

	var cropX int
	// if boundsHeight > cropSize.Height || hGravity == "center" {
	if hGravity == "center" {
		cropX = (boundsX + int(math.Round(float64(boundsWidth)/2.0))) - (cropSize.Width / 2)
	} else if hGravity == "left" {
		cropX = boundsX - actualPadding
	} else {
		cropX = ((boundsX + boundsWidth) + actualPadding) - cropSize.Width
	}
	cropX = utils.Min(sourceWidth-cropSize.Width, utils.Max(0, cropX))

	var cropY int
	// if boundsHeight > cropSize.Height || vGravity == "center" {
	if vGravity == "center" {
		cropY = (boundsY + int(math.Round(float64(boundsHeight)/2.0))) - (cropSize.Height / 2)
	} else if vGravity == "top" {
		cropY = boundsY - actualPadding
	} else {
		cropY = ((boundsY + boundsHeight) + actualPadding) - cropSize.Height
	}
	cropY = utils.Min(sourceHeight-cropSize.Height, utils.Max(0, cropY))

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width != targetCropWidth) || (cropSize.Height != targetCropHeight) {
		err = sourceImage.ThumbnailWithSize(targetCropWidth, targetCropHeight, vips.InterestingNone, vips.SizeBoth)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}
