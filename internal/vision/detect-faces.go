package vision

import (
	"foxy/internal/config"
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
)

func DetectFaces(sourceConfig config.Config, sid string, key string, sourceImage *vips.ImageRef, skipCache bool) (*metadata.Metadata, error) {
	if !sourceConfig.Vision.Enabled {
		return nil, nil
	}

	if sourceConfig.Vision.Type == "rekognition" {
		return RekognitionDetectFaces(sourceConfig, sid, key, sourceImage, skipCache)
	} else {
		return nil, nil
	}
}
