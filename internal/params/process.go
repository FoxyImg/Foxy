package params

import (
	"errors"
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
		foundMeta, err := vision.DetectFaces(sourceId, sourceConfig, sid, key, sourceImage, params.Debug != nil && params.Debug.DisableMetaCache)
		if err != nil {
			return nil, nil, err
		}

		if foundMeta == nil {
			return nil, nil, errors.New("meta is nil?")
		}

		visionMeta = foundMeta
		visionMeta.Width = sourceImage.Width()
		visionMeta.Height = sourceImage.Height()

		if params.MetaOnly {
			return nil, visionMeta, nil
		}
	} else {
		visionMeta = &vision.Metadata{
			Width:  sourceImage.Width(),
			Height: sourceImage.Height(),
		}
	}

	var err error

	if params.BackgroundRemoval != nil {
		sourceImage, err = params.BackgroundRemoval.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.SourceCrop != nil {
		sourceImage, err = params.SourceCrop.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Redact != nil {
		sourceImage, err = params.Redact.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Debug != nil {
		_, _ = params.Debug.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
	}

	if params.Size != nil {
		sourceImage, err = params.Size.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Rotation != nil {
		sourceImage, err = params.Rotation.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Adjustments != nil {
		sourceImage, err = params.Adjustments.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.GradientMap != nil {
		sourceImage, err = params.GradientMap.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Stylize != nil {
		sourceImage, err = params.Stylize.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(params.Overlays) > 0 {
		sourceImage, err = params.Overlays.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Padding != nil {
		sourceImage, err = params.Padding.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Border != nil {
		sourceImage, err = params.Border.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	if params.Mask != nil {
		sourceImage, err = params.Mask.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if err != nil {
			return nil, nil, err
		}
	}

	buffer, err := params.Export.Export(sourceImage, params)
	if err != nil {
		return nil, nil, err
	}

	return buffer, visionMeta, nil
}
