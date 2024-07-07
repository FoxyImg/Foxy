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
			sourceImage.Close()
			return nil, nil, err

		}

		newImageBuffer, _, err := sourceImage.ExportNative()
		if err != nil {
			return nil, nil, err
		}

		sourceImage.Close()

		newImage, err := vips.NewImageFromBuffer(newImageBuffer)
		if err != nil {
			if newImage != nil {
				newImage.Close()
			}
			return nil, nil, err
		}

		sourceImage = newImage
	}

	var visionMeta *vision.Metadata
	if params.NeedsVision && sourceConfig.Vision.Enabled {
		foundMeta, err := vision.DetectFaces(sourceId, sourceConfig, sid, key, sourceImage, params.Debug != nil && params.Debug.DisableMetaCache)
		if err != nil {
			sourceImage.Close()
			return nil, nil, err
		}

		if foundMeta == nil {
			sourceImage.Close()
			return nil, nil, errors.New("meta is nil?")
		}

		visionMeta = foundMeta
		visionMeta.Width = sourceImage.Width()
		visionMeta.Height = sourceImage.Height()

		if params.MetaOnly {
			sourceImage.Close()
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
		processedImage, processErr := params.BackgroundRemoval.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.SourceCrop != nil {
		processedImage, processErr := params.SourceCrop.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Redact != nil {
		processedImage, processErr := params.Redact.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Debug != nil {
		_, _ = params.Debug.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
	}

	if params.Sizing != nil {
		processedImage, processErr := params.Sizing.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Rotation != nil {
		processedImage, processErr := params.Rotation.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Levels != nil {
		processedImage, processErr := params.Levels.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Adjustments != nil {
		processedImage, processErr := params.Adjustments.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.GradientMap != nil {
		processedImage, processErr := params.GradientMap.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Stylize != nil {
		processedImage, processErr := params.Stylize.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if len(params.Overlays) > 0 {
		processedImage, processErr := params.Overlays.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Padding != nil {
		processedImage, processErr := params.Padding.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Border != nil {
		processedImage, processErr := params.Border.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Mask != nil {
		processedImage, processErr := params.Mask.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	buffer, err := params.Export.Export(sourceImage, params)
	sourceImage.Close()
	if err != nil {
		return nil, nil, err
	}

	return buffer, visionMeta, nil
}
