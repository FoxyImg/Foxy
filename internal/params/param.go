package params

import (
	"foxy/internal/config"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
)

type Param interface {
	Params() []string
	ParseParams(param string, options []string) (needsVision bool)
}

type ProcessingParam interface {
	Param
	Process(sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error)
}
