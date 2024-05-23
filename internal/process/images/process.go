package images

import (
	"foxy/internal/aws"
	"foxy/internal/config"
	"foxy/internal/metadata"
	. "foxy/internal/process/images/crop"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"time"
)

func ProcessImage(
	sourceConfig config.Config,
	sid string,
	key string,
	params *metadata.ImageParams,
	sourceImage *vips.ImageRef,
) (*[]byte, *metadata.Metadata, error) {
	defer utils.TrackTime(time.Now(), "Process Image")

	var imageMeta *metadata.Metadata
	if params.NeedsVision && sourceConfig.Vision.Enabled {
		//TODO: Fix obvious race condition
		if sourceConfig.Vision.Type == "rekognition" {
			foundMeta, err := aws.DetectFaces(sourceConfig, sid, key, sourceImage, params.Debug.DisableMetaCache)
			if err != nil {
				return nil, nil, err
			}

			imageMeta = foundMeta
		}

		if params.MetaOnly {
			return nil, imageMeta, nil
		}
	}

	if (params.Debug.Faces || params.Debug.AllFaces || params.Debug.People || params.Debug.AllPeople || params.Debug.OtherLabels) && imageMeta != nil {
		drawDebugBounds(imageMeta, params, sourceImage)
	}

	sourceImage, err := Crop(imageMeta, params, sourceImage)
	if err != nil {
		return nil, nil, err
	}

	if params.Blur != nil {
		err := sourceImage.GaussianBlur(float64(*params.Blur))
		if err != nil {
			return nil, nil, err
		}
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

		return &buffer, imageMeta, nil
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

		return &buffer, imageMeta, nil
	} else if params.ExportParams.Format == "avif" {
		avif := vips.NewAvifExportParams()
		avif.Quality = params.ExportParams.Quality
		avif.StripMetadata = true

		buffer, _, err := sourceImage.ExportAvif(avif)
		if err != nil {
			log.Println("Export AVIF error:", err)
			return nil, nil, err
		}

		return &buffer, imageMeta, nil
	} else {
		jpg := vips.NewJpegExportParams()
		jpg.Quality = params.ExportParams.Quality
		jpg.StripMetadata = true

		buffer, _, err := sourceImage.ExportJpeg(jpg)
		if err != nil {
			log.Println("Export JPEG error:", err)
			return nil, nil, err
		}

		return &buffer, imageMeta, nil
	}
}
