package utils

import (
	"github.com/davidbyttow/govips/v2/vips"
	"time"
)

func RenderImage(sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	defer TrackTime(time.Now(), "Render Image")

	png, _, err := sourceImage.ExportPng(nil)
	if err != nil {
		return sourceImage, err
	}

	pngImage, err := vips.NewImageFromBuffer(png)
	if err != nil {
		return sourceImage, err
	}

	return pngImage, nil
}
