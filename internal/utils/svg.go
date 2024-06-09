package utils

import (
	"github.com/davidbyttow/govips/v2/vips"
	"time"
)

func RenderSVG(svgString string) (*vips.ImageRef, error) {
	defer TrackTime(time.Now(), "Render SVG")

	svg, err := vips.LoadImageFromBuffer([]byte(svgString), vips.NewImportParams())
	if err != nil {
		return nil, err
	}

	buffer, _, err := svg.ExportPng(nil)
	if err != nil {
		return nil, err
	}

	return vips.NewImageFromBuffer(buffer)
}
