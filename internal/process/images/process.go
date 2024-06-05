package images

import (
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"time"
)

func ProcessImage(
	sourceId string,
	sourceConfig *config.Config,
	sid string,
	key string,
	params *ImageParams,
	sourceImage *vips.ImageRef,
) (*[]byte, *vision.Metadata, error) {
	defer utils.TrackTime(time.Now(), "Process Image")

	if sourceImage.Width() > env.FoxyEnvironment.MaxSourceSize || sourceImage.Height() > env.FoxyEnvironment.MaxSourceSize {
		scale := float64(env.FoxyEnvironment.MaxSourceSize) / float64(utils.Max(sourceImage.Width(), sourceImage.Height()))
		err := sourceImage.Resize(scale, vips.KernelLanczos3)
		if err != nil {
			return nil, nil, err
		}

		newImageBuffer, _, err := sourceImage.ExportNative()
		if err != nil {
			return nil, nil, err
		}

		sourceImage, err = vips.NewImageFromBuffer(newImageBuffer)
		if err != nil {
			return nil, nil, err
		}
	}

	var visionMeta *vision.Metadata
	if params.NeedsVision && sourceConfig.Vision.Enabled {
		foundMeta, err := vision.DetectFaces(sourceId, sourceConfig, sid, key, sourceImage, params.Debug.DisableMetaCache)
		if err != nil {
			return nil, nil, err
		}

		visionMeta = foundMeta

		if params.MetaOnly {
			return nil, visionMeta, nil
		}
	}

	var err error

	sourceImage, err = params.Redact.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	_, _ = params.Debug.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)

	sourceImage, err = params.Size.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Rotation.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Adjustments.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.GradientMap.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Stylize.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Watermark.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Padding.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Border.Process(sourceId, sourceConfig, sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	buffer, err := params.Export.Export(sourceImage)
	if err != nil {
		return nil, nil, err
	}

	return buffer, visionMeta, nil
}
