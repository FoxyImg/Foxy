package images

import (
	"foxy/internal/config"
	"foxy/internal/env"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"time"
)

func ProcessImage(
	sourceConfig config.Config,
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
		foundMeta, err := vision.DetectFaces(sourceConfig, sid, key, sourceImage, params.Debug.DisableMetaCache)
		if err != nil {
			return nil, nil, err
		}

		visionMeta = foundMeta

		if params.MetaOnly {
			return nil, visionMeta, nil
		}
	}

	var err error

	sourceImage, err = params.Redact.Process(sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	_, _ = params.Debug.Process(sourceImage, params, visionMeta)

	sourceImage, err = params.Size.Process(sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Stylize.Process(sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Padding.Process(sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	sourceImage, err = params.Border.Process(sourceImage, params, visionMeta)
	if err != nil {
		return nil, nil, err
	}

	if params.ExportParams.Format == "png" {
		png := vips.NewPngExportParams()
		png.Quality = params.ExportParams.Quality
		png.StripMetadata = true

		buffer, _, err := sourceImage.ExportPng(png)
		if err != nil {
			log.Println("Export PNG Error:", err)
			return nil, nil, err
		}

		return &buffer, visionMeta, nil
	} else if params.ExportParams.Format == "webp" {
		webp := vips.NewWebpExportParams()
		webp.Quality = params.ExportParams.Quality
		webp.StripMetadata = true
		if params.ExportParams.Lossless != nil {
			webp.Lossless = *params.ExportParams.Lossless
		}
		if params.ExportParams.NearLossless != nil {
			webp.NearLossless = *params.ExportParams.NearLossless
		}
		if params.ExportParams.ReductionEffort != nil {
			webp.ReductionEffort = *params.ExportParams.ReductionEffort
		}

		buffer, _, err := sourceImage.ExportWebp(webp)
		if err != nil {
			log.Println("Export WebP Error:", err)
			return nil, nil, err
		}

		return &buffer, visionMeta, nil
	} else if params.ExportParams.Format == "avif" {
		avif := vips.NewAvifExportParams()
		avif.Quality = params.ExportParams.Quality
		avif.StripMetadata = true

		buffer, _, err := sourceImage.ExportAvif(avif)
		if err != nil {
			log.Println("Export AVIF error:", err)
			return nil, nil, err
		}

		return &buffer, visionMeta, nil
	} else {
		jpg := vips.NewJpegExportParams()
		jpg.Quality = params.ExportParams.Quality
		jpg.StripMetadata = true

		buffer, _, err := sourceImage.ExportJpeg(jpg)
		if err != nil {
			log.Println("Export JPEG error:", err)
			return nil, nil, err
		}

		return &buffer, visionMeta, nil
	}
}
