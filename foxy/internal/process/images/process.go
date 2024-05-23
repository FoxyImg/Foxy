package images

import (
	"fmt"
	"foxy/internal/aws"
	"foxy/internal/db"
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"strconv"
	"time"
)

func drawBounds(label string, image *vips.ImageRef, box geometry.Box, color vips.ColorRGBA, fill bool) {
	log.Println("Draw Rect:", box)
	sw := image.Width()
	sh := image.Height()

	bx := int(math.Floor(box.Left * float64(sw)))
	bt := int(math.Floor(box.Top * float64(sh)))
	bw := int(math.Floor(box.Width * float64(sw)))
	bh := int(math.Floor(box.Height * float64(sh)))

	svgRect := fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" style=\"fill:transparent; stroke:rgba(%d, %d, %d, %f); stroke-width:16\" />",
		bx,
		bt,
		bw,
		bh,
		color.R,
		color.G,
		color.B,
		float64(color.A)/255.0,
	)

	svgText := fmt.Sprintf("<text x=\"%d\" y=\"%d\" style=\"fill:white; font-size: 120pt; paint-order: stroke; stroke: rgba(%d, %d, %d, %f); stroke-width: 8px; stroke-linecap: butt; stroke-linejoin: miter; font-weight: 800;\">%s</text>",
		bx,
		bt,
		color.R,
		color.G,
		color.B,
		float64(color.A)/255.0,
		label,
	)

	svgString := fmt.Sprintf("<svg width=\"%d\" height=\"%d\" xmlns=\"http://www.w3.org/2000/svg\">%s%s</svg>",
		sw,
		sh,
		svgRect,
		svgText,
	)

	fmt.Println(svgString)
	svg, err := vips.LoadImageFromBuffer([]byte(svgString), vips.NewImportParams())
	if err != nil {
		log.Println("Draw Rect Error:", err)
	}

	err = image.Composite(svg, vips.BlendModeOver, 0, 0)
	if err != nil {
		log.Println("Draw Rect Error:", err)
	}
}

func cropFace(imageMeta *metadata.Metadata, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var faceBounds geometry.Box
	if params.FaceIndex != nil && *params.FaceIndex < len(imageMeta.Faces) {
		faceBounds = imageMeta.Faces[*params.FaceIndex].Box
	} else {
		faceBounds = metadata.CalcFacesBounds(imageMeta.Faces)
	}

	fy := int(math.Floor(faceBounds.Top * float64(sh)))
	fh := int(math.Floor(faceBounds.Height * float64(sh)))

	var targetWidth int
	var targetHeight int

	if params.Width == nil {
		targetWidth = 10000
	} else {
		targetWidth = *params.Width
	}

	if params.Height == nil {
		targetHeight = 10000
	} else {
		targetHeight = *params.Height
	}

	cropSize := geometry.SizeToFitSize(targetWidth, targetHeight, sw, sh)

	cx := int((faceBounds.Left + (faceBounds.Width / 2.0)) * float64(sw))
	var cy int
	if fh > cropSize.Height {
		cy = fy + int(math.Floor(float64(fh)/2.0))
	} else {
		cy = fy - int(math.Floor(48.0*(float64(targetWidth)/1600.0)))
	}

	cropX := utils.Min(sw, utils.Max(0, cx-(cropSize.Width/2)))
	var cropY int
	if fh > cropSize.Height {
		cropY = utils.Min(sw, utils.Max(0, cy-(cropSize.Height/2)))
	} else {
		cropY = utils.Min(sh, utils.Max(0, cy))
	}

	if (cropX + cropSize.Width) > sw {
		cropX = utils.Max(0, sw-cropSize.Width)
	}

	if (cropY + cropSize.Height) > sh {
		cropY = utils.Max(0, sh-cropSize.Height)
	}

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width > targetWidth) || (cropSize.Height > targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}

func cropPerson(imageMeta *metadata.Metadata, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var faceBounds geometry.Box

	var personLabels = metadata.GetPersonLabels(imageMeta.Labels)
	if params.PersonIndex != nil && *params.PersonIndex < len(personLabels) {
		faceBounds = *personLabels[*params.PersonIndex].Box
	} else {
		faceBounds = metadata.CalcPersonBounds(imageMeta.Labels)
	}

	var targetWidth int
	var targetHeight int

	if params.Width == nil {
		targetWidth = 10000
	} else {
		targetWidth = *params.Width
	}

	if params.Height == nil {
		targetHeight = 10000
	} else {
		targetHeight = *params.Height
	}

	cropSize := geometry.SizeToFitSize(targetWidth, targetHeight, sw, sh)

	cx := int((faceBounds.Left + (faceBounds.Width / 2.0)) * float64(sw))
	cy := int((faceBounds.Top + (faceBounds.Height / 2.0)) * float64(sh))

	cropX := utils.Min(sw, utils.Max(0, cx-(cropSize.Width/2)))
	cropY := utils.Min(sh, utils.Max(0, cy-(cropSize.Height/2)))

	if (cropX + cropSize.Width) > sw {
		cropX = utils.Max(0, sw-cropSize.Width)
	}

	if (cropY + cropSize.Height) > sh {
		cropY = utils.Max(0, sh-cropSize.Height)
	}

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width > targetWidth) || (cropSize.Height > targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}

func cropFit(params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	return sourceImage, nil
}

func cropFill(params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	err := sourceImage.ThumbnailWithSize(*params.Width, *params.Height, vips.InterestingCentre, vips.SizeBoth)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}

func resize(params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var targetWidth int
	var targetHeight int

	if params.Width == nil {
		targetWidth = 10000
	} else {
		targetWidth = *params.Width
	}

	if params.Height == nil {
		targetHeight = 10000
	} else {
		targetHeight = *params.Height
	}

	err := sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeBoth)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}

func ProcessImage(
	sourceConfig db.SourceConfig,
	sid string,
	key string,
	params *ImageParams,
	sourceImage *vips.ImageRef,
) (*[]byte, *metadata.Metadata, error) {
	defer utils.TrackTime(time.Now(), "Process Image")

	var imageMeta *metadata.Metadata
	if params.NeedsVision {
		meta, err := aws.DetectFaces(sourceConfig, sid, key, sourceImage, params.DisableMetaCache)
		if err != nil {
			return nil, nil, err
		}

		if params.MetaOnly {
			return nil, meta, nil
		}

		imageMeta = meta
	}

	if (params.DebugFaces || params.DebugAllFaces || params.DebugPeople || params.DebugAllPeople || params.DebugOtherLabels) && imageMeta != nil {
		if params.DebugOtherLabels {
			otherLabel := metadata.FilterNonPersonLabels(imageMeta.Labels)
			for _, label := range otherLabel {
				drawBounds(label.Name, sourceImage, *label.Box, vips.ColorRGBA{R: 255, G: 0, B: 255, A: 255}, false)
			}
		}

		if params.DebugAllPeople && len(imageMeta.People) > 0 {
			drawBounds("all people", sourceImage, metadata.CalcLabelsBounds(imageMeta.People), vips.ColorRGBA{R: 0, G: 255, B: 255, A: 255}, false)
		}

		if params.DebugPeople && len(imageMeta.People) > 0 {
			for idx, label := range imageMeta.People {
				if label.Box != nil {
					log.Println("Person #", idx, label.Name)
					drawBounds("#"+strconv.Itoa(idx)+" - "+label.Name, sourceImage, *label.Box, vips.ColorRGBA{R: 0, G: 255, B: 0, A: 255}, false)
				}
			}
		}

		if params.DebugAllFaces && len(imageMeta.Faces) > 0 {
			drawBounds("all faces", sourceImage, metadata.CalcFacesBounds(imageMeta.Faces), vips.ColorRGBA{R: 255, G: 255, B: 0, A: 255}, false)
		}

		if params.DebugFaces && len(imageMeta.Faces) > 0 {
			for idx, face := range imageMeta.Faces {
				drawBounds("face #"+strconv.Itoa(idx), sourceImage, face.Box, vips.ColorRGBA{R: 255, G: 0, B: 0, A: 255}, false)
			}
		}
	}

	if params.Width != nil || params.Height != nil {
		if params.Crop != nil {
			for _, cropMode := range *params.Crop {
				if cropMode == "face" && imageMeta != nil && len(imageMeta.Faces) > 0 && params.Width != nil && params.Height != nil {
					croppedImage, err := cropFace(imageMeta, params, sourceImage)
					if err != nil {
						return nil, nil, err
					}

					sourceImage = croppedImage
				} else if cropMode == "person" && imageMeta != nil && len(imageMeta.Labels) > 0 && metadata.HasPersonLabel(imageMeta.Labels) && params.Width != nil && params.Height != nil {
					croppedImage, err := cropPerson(imageMeta, params, sourceImage)
					if err != nil {
						return nil, nil, err
					}

					sourceImage = croppedImage
				} else if cropMode == "fit" && params.Width != nil && params.Height != nil {
					croppedImage, err := cropFit(params, sourceImage)
					if err != nil {
						return nil, nil, err
					}

					sourceImage = croppedImage
				} else if cropMode == "fill" && params.Width != nil && params.Height != nil {
					croppedImage, err := cropFill(params, sourceImage)
					if err != nil {
						return nil, nil, err
					}

					sourceImage = croppedImage
				} else {
					//return nil, nil, errors.New("unknown crop mode")
				}
			}
		} else {
			resizedImage, err := resize(params, sourceImage)
			if err != nil {
				return nil, nil, err
			}

			sourceImage = resizedImage
		}
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
