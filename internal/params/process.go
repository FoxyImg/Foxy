package params

import (
	"errors"
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"time"
)

//type processData struct {
//	SourceConfig *config.Config `json:"sourceConfig"`
//	SourceId     string         `json:"sourceId"`
//	Sid          string         `json:"sid"`
//	Key          string         `json:"key"`
//	Params       *ImageParams   `json:"params"`
//}

func ProcessImage(
	sourceId string,
	sourceConfig *config.Config,
	sid string,
	key string,
	params *ImageParams,
	sourceImage *vips.ImageRef,
) (*[]byte, *vision.Metadata, error) {
	defer utils.TrackTime(time.Now(), "Process Image")

	//paramsJSON, jsonErr := json.MarshalIndent(processData{
	//	SourceConfig: sourceConfig,
	//	SourceId:     sourceId,
	//	Sid:          sid,
	//	Key:          key,
	//	Params:       params,
	//}, "", "  ")
	//if jsonErr != nil {
	//	return nil, nil, jsonErr
	//}
	//_ = os.WriteFile("./process-data/"+strings.ReplaceAll(key, "/", "-")+".json", paramsJSON, 0644)

	if sourceImage.Width() > env.FoxyEnvironment.MaxSourceSize || sourceImage.Height() > env.FoxyEnvironment.MaxSourceSize {
		log.Println("Resize Image", key)
		scale := float64(env.FoxyEnvironment.MaxSourceSize) / float64(utils.Max(sourceImage.Width(), sourceImage.Height()))
		err := sourceImage.Resize(scale, vips.KernelLanczos3)
		if err != nil {
			log.Println("Resize Image Err", key, err)
			sourceImage.Close()
			return nil, nil, err
		}

		log.Println("Export Resize Image Buffer", key)
		newImageBuffer, _, err := sourceImage.ExportPng(nil)
		if err != nil {
			log.Println("Export Resize Image Buffer Error", key, err)
			return nil, nil, err
		}

		sourceImage.Close()

		log.Println("Load Image From Buffer", key)
		newImage, err := vips.NewImageFromBuffer(newImageBuffer)
		if err != nil {
			log.Println("Load Image From Buffer Error", key, err)
			if newImage != nil {
				newImage.Close()
			}
			return nil, nil, err
		}

		sourceImage = newImage
	}

	var visionMeta *vision.Metadata
	if params.NeedsVision && sourceConfig.Vision.Enabled {
		log.Println("Detecting faces", key)
		foundMeta, err := vision.DetectFaces(sourceId, sourceConfig, sid, key, sourceImage, params.Debug != nil && params.Debug.DisableMetaCache)
		if err != nil {
			log.Println("Detecting faces error", key, err)
			sourceImage.Close()
			return nil, nil, err
		}

		if foundMeta == nil {
			log.Println("Detecting faces - no found meta", key)
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
		log.Println("Background Removal", key)
		processedImage, processErr := params.BackgroundRemoval.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Background Removal Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Background Removal - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.SourceCrop != nil {
		log.Println("Source Crop", key)
		processedImage, processErr := params.SourceCrop.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Source Crop Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Source Crop - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Redact != nil {
		log.Println("Redact", key)
		processedImage, processErr := params.Redact.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Redact Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Redact - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Debug != nil {
		log.Println("Debug", key)
		_, _ = params.Debug.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
	}

	if params.Sizing != nil {
		log.Println("Sizing", key)
		processedImage, processErr := params.Sizing.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Sizing Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Sizing - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Rotation != nil {
		log.Println("Rotation", key)
		processedImage, processErr := params.Rotation.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Rotation Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Rotation - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Levels != nil {
		log.Println("Levels", key)
		processedImage, processErr := params.Levels.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Levels Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Levels - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Adjustments != nil {
		log.Println("Adjustments", key)
		processedImage, processErr := params.Adjustments.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Adjustments Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Adjustments - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.GradientMap != nil {
		log.Println("Gradient Map", key)
		processedImage, processErr := params.GradientMap.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Gradient Map Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Gradient Map - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Stylize != nil {
		log.Println("Stylize", key)
		processedImage, processErr := params.Stylize.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Stylize Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Stylize - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if len(params.Overlays) > 0 {
		log.Println("Overlays", key)
		processedImage, processErr := params.Overlays.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Overlays Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Overlays - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Padding != nil {
		log.Println("Padding", key)
		processedImage, processErr := params.Padding.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Padding Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Padding - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Border != nil {
		log.Println("Border", key)
		processedImage, processErr := params.Border.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Border Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Border - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	if params.Mask != nil {
		log.Println("Mask", key)
		processedImage, processErr := params.Mask.Process(key, sourceId, sourceConfig, sourceImage, params, visionMeta)
		if processErr != nil {
			log.Println("Mask Err", key, err)
			if processedImage != nil {
				processedImage.Close()
			}
			return nil, nil, processErr
		}

		if processedImage != sourceImage {
			log.Println("Mask - Changed Source", key)
			sourceImage.Close()
			sourceImage = processedImage
		}
	}

	log.Println("Export", key)
	buffer, err := params.Export.Export(sourceImage, params)
	sourceImage.Close()
	if err != nil {
		log.Println("Export Err", key, err)
		return nil, nil, err
	}

	log.Println("Done Processing", key)
	return buffer, visionMeta, nil
}
