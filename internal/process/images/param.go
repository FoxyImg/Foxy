package images

import (
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
)

type Param interface {
	Params() []string
	ParseParams(param string, options []string) (needsVision bool)
}

type ProcessingParam interface {
	Param
	Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error)
}
