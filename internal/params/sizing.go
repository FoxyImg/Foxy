package params

import (
	"errors"
	"foxy/internal/config"
	"foxy/internal/geometry"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"slices"
	"strconv"
	"time"
)

type BoundingBoxCropParams struct {
	Index    *int     `json:"index,omitempty"`
	Padding  *int     `json:"padding,omitempty"`
	Zoom     *float64 `json:"zoom,omitempty"`
	HGravity *string  `json:"hGravity,omitempty"`
	VGravity *string  `json:"vGravity,omitempty"`
	Largest  *bool    `json:"largest,omitempty"`
	Smallest *bool    `json:"smallest,omitempty"`
	Focus    *bool    `json:"focus,omitempty"`
}

type FocalPointOptions struct {
	X    *float64 `json:"x,omitempty"`
	Y    *float64 `json:"y,omitempty"`
	Zoom *float64 `json:"zoom,omitempty"`
}

type SizingOptions struct {
	Crop        *[]string              `json:"crop,omitempty"`
	Width       *int                   `json:"width,omitempty"`
	Height      *int                   `json:"height,omitempty"`
	AspectRatio *float64               `json:"aspectRatio,omitempty"`
	Zoom        *float64               `json:"zoom,omitempty"`
	HGravity    *string                `json:"hGravity,omitempty"`
	VGravity    *string                `json:"vGravity,omitempty"`
	Face        *BoundingBoxCropParams `json:"face,omitempty"`
	Person      *BoundingBoxCropParams `json:"person,omitempty"`
	Interesting *vips.Interesting      `json:"interesting,omitempty"`
	FocalPoint  *FocalPointOptions     `json:"focalPoint,omitempty"`
}

func (*SizingOptions) Params() []string {
	return []string{"crop", "w", "h", "d", "zoom", "ar", "fp", "face", "person", "smart", "gravity"}
}

func (bbx *BoundingBoxCropParams) parseBoundingBoxCropParams(options []string) {
	if len(options) < 1 {
		return
	}

	if options[0] == "index" && len(options) == 2 {
		t := true
		if options[1] == "largest" {
			bbx.Largest = &t
		} else if options[1] == "smallest" {
			bbx.Smallest = &t
		} else {
			f, err := strconv.Atoi(options[1])
			if err == nil {
				bbx.Index = &f
			}
		}
	} else if options[0] == "pad" && len(options) == 2 {
		f, err := strconv.Atoi(options[1])
		if err == nil {
			bbx.Padding = &f
		}
	} else if options[0] == "zoom" && len(options) == 2 {
		f, err := strconv.ParseFloat(options[1], 64)
		if err == nil {
			f = f / 100.0
			bbx.Zoom = &f
		}
	} else if options[0] == "gravity" && len(options) == 3 {
		bbx.HGravity = &options[1]
		bbx.VGravity = &options[2]
	} else if options[0] == "focus" {
		bbx.Focus = utils.Ptr(true)
	}
}

func (sz *SizingOptions) ParseParams(param string, options []string) (needsVision bool) {
	needsVision = false
	switch param {
	case "crop":
		sz.Crop = &options
		needsVision = slices.Contains(*sz.Crop, "face") || slices.Contains(*sz.Crop, "person")
	case "w":
		if len(options) != 1 {
			return
		}

		w, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		sz.Width = &w
	case "h":
		if len(options) != 1 {
			return
		}

		h, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		sz.Height = &h
	case "d":
		if len(options) != 2 {
			return
		}

		sz.Width, sz.Height = Int2DVectorVal(options)
	case "zoom":
		if len(options) != 1 {
			return
		}

		z, err := strconv.ParseFloat(options[0], 64)
		if err != nil {
			return
		}

		sz.Zoom = &z
	case "ar":
		if len(options) != 2 {
			return
		}

		arw, err := strconv.Atoi(options[0])
		if err != nil {
			return
		}

		arh, err := strconv.Atoi(options[1])
		if err != nil {
			return
		}

		ar := float64(arw) / float64(arh)
		sz.AspectRatio = &ar
	case "fp":
		if len(options) < 1 {
			return
		}

		if sz.FocalPoint == nil {
			sz.FocalPoint = &FocalPointOptions{}
		}

		if options[0] == "zoom" {
			if len(options) != 2 {
				return
			}

			z, err := strconv.ParseFloat(options[1], 64)
			if err == nil && z > 0 {
				z = z / 100.0
				sz.FocalPoint.Zoom = &z
			}
		} else if len(options) == 2 {
			fpx, err := strconv.ParseFloat(options[0], 64)
			if err != nil {
				return
			}

			fpy, err := strconv.ParseFloat(options[1], 64)
			if err != nil {
				return
			}

			sz.FocalPoint.X = utils.Ptr(fpx)
			sz.FocalPoint.Y = utils.Ptr(fpy)
		}
	case "face":
		if sz.Face == nil {
			sz.Face = &BoundingBoxCropParams{}
		}
		needsVision = true
		sz.Face.parseBoundingBoxCropParams(options)
	case "person":
		if sz.Person == nil {
			sz.Person = &BoundingBoxCropParams{}
		}
		needsVision = true
		sz.Person.parseBoundingBoxCropParams(options)
	case "smart":
		if len(options) != 1 {
			return
		}

		var interesting vips.Interesting
		switch options[0] {
		case "entropy":
			interesting = vips.InterestingEntropy
		case "high":
			interesting = vips.InterestingHigh
		case "low":
			interesting = vips.InterestingLow
		case "center":
			interesting = vips.InterestingCentre
		case "centre":
			interesting = vips.InterestingCentre
		case "none":
			interesting = vips.InterestingNone
		default:
			interesting = vips.InterestingAttention
		}

		sz.Interesting = &interesting
	case "gravity":
		if len(options) == 1 {
			sz.HGravity = &options[0]
			sz.VGravity = &options[0]
		} else if len(options) == 2 {
			sz.HGravity = &options[0]
			sz.VGravity = &options[1]
		}
	}

	return
}

func (sz *SizingOptions) Process(sourceKey string, sourceId string, config *config.Config, sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if sz.Width != nil || sz.Height != nil {
		defer utils.TrackTime(time.Now(), "Sizing")

		cW := sz.Width
		cH := sz.Height

		if sz.AspectRatio != nil {
			if cW != nil {
				nh := int(math.Round(float64(*cW) / *sz.AspectRatio))
				cH = &nh
			} else if cH != nil {
				nw := int(math.Round(float64(*cH) * *sz.AspectRatio))
				cW = &nw
			}
		}

		if sz.Crop != nil {
			for _, cropMode := range *sz.Crop {
				if cropMode == "face" && sz.Face != nil && imageMeta != nil && len(imageMeta.Faces) > 0 && cW != nil && cH != nil {
					croppedImage, err := sz.cropFace(cW, cH, imageMeta, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "person" && sz.Person != nil && imageMeta != nil && len(imageMeta.People) > 0 && cW != nil && cH != nil {
					croppedImage, err := sz.cropPerson(cW, cH, imageMeta, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "fit" && cW != nil && cH != nil {
					croppedImage, err := sz.cropFit(cW, cH, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "smart" && cW != nil && cH != nil {
					croppedImage, err := sz.cropSmart(cW, cH, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "crop" && cW != nil && cH != nil {
					croppedImage, err := sz.cropFill(cW, cH, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "focus" && sz.FocalPoint != nil && cW != nil && cH != nil && sz.FocalPoint.X != nil && sz.FocalPoint.Y != nil {
					croppedImage, err := sz.cropFocus(*sz.FocalPoint.X, *sz.FocalPoint.Y, cW, cH, params, sourceImage, sz.FocalPoint.Zoom)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				}
			}

			if cW != nil && cH != nil {
				croppedImage, err := sz.cropFill(cW, cH, params, sourceImage)
				if err != nil {
					return nil, err
				}

				return croppedImage, nil
			} else if cW != nil || cH != nil {
				resizedImage, err := sz.resize(params, sourceImage)
				if err != nil {
					return nil, err
				}

				return resizedImage, nil
			} else {
				return nil, errors.New("unknown crop mode")
			}
		} else {
			resizedImage, err := sz.resize(params, sourceImage)
			if err != nil {
				return nil, err
			}

			return resizedImage, nil
		}
	}

	return sourceImage, nil
}

func (sz *SizingOptions) cropBounds(cW *int, cH *int, bounds geometry.Box, cropZoom *float64, boundsZoom *float64, padding int, hGravity string, vGravity string, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sourceWidth := sourceImage.Width()
	sourceHeight := sourceImage.Height()

	if bounds.Width*bounds.Height == 0 {
		return sz.cropFill(cW, cH, params, sourceImage)
	}

	boundsX := int(math.Round(bounds.Left * float64(sourceWidth)))
	boundsY := int(math.Round(bounds.Top * float64(sourceHeight)))
	boundsWidth := int(math.Round(bounds.Width * float64(sourceWidth)))
	boundsHeight := int(math.Round(bounds.Height * float64(sourceHeight)))

	var targetCropWidth int
	var targetCropHeight int

	if cW == nil {
		targetCropWidth = 10000
	} else {
		targetCropWidth = *cW
	}

	if cH == nil {
		targetCropHeight = 10000
	} else {
		targetCropHeight = *cH
	}

	cropSize := geometry.SizeToFitSize(targetCropWidth, targetCropHeight, sourceWidth, sourceHeight)

	actualPadding := int(math.Round((float64(padding) / 512.0) * float64(sourceWidth)))
	var zoom *float64 = nil
	if boundsZoom != nil {
		z := math.Min(float64(cropSize.Height)/float64(boundsHeight+(actualPadding*2)), float64(cropSize.Width)/float64(boundsWidth+(actualPadding*2))) * *boundsZoom
		zoom = &z
	} else if cropZoom != nil {
		zoom = cropZoom
	}

	if zoom != nil {
		newCW := int(math.Round(float64(cropSize.Width) * (1 / *zoom)))
		newCH := int(math.Round(float64(cropSize.Height) * (1 / *zoom)))
		if newCW <= sourceWidth && newCH <= sourceHeight {
			cropSize.Width = newCW
			cropSize.Height = newCH
		}
	}

	var cropX int
	// if boundsHeight > cropSize.Height || hGravity == "center" {
	if hGravity == "center" {
		cropX = (boundsX + int(math.Round(float64(boundsWidth)/2.0))) - (cropSize.Width / 2)
	} else if hGravity == "left" {
		cropX = boundsX - actualPadding
	} else {
		cropX = ((boundsX + boundsWidth) + actualPadding) - cropSize.Width
	}
	cropX = utils.Min(sourceWidth-cropSize.Width, utils.Max(0, cropX))

	var cropY int
	// if boundsHeight > cropSize.Height || vGravity == "center" {
	if vGravity == "center" {
		cropY = (boundsY + int(math.Round(float64(boundsHeight)/2.0))) - (cropSize.Height / 2)
	} else if vGravity == "top" {
		cropY = boundsY - actualPadding
	} else {
		cropY = ((boundsY + boundsHeight) + actualPadding) - cropSize.Height
	}
	cropY = utils.Min(sourceHeight-cropSize.Height, utils.Max(0, cropY))

	err := sourceImage.Crop(
		cropX,
		cropY,
		cropSize.Width,
		cropSize.Height,
	)
	if err != nil {
		return nil, err
	}

	if (cropSize.Width != targetCropWidth) || (cropSize.Height != targetCropHeight) {
		err = sourceImage.ThumbnailWithSize(targetCropWidth, targetCropHeight, vips.InterestingNone, vips.SizeBoth)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}

func (sz *SizingOptions) cropFace(cW *int, cH *int, imageMeta *vision.Metadata, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	if sz.Face == nil {
		return sz.cropFill(cW, cH, params, sourceImage)
	}

	var faceBounds geometry.Box
	if sz.Face.Index != nil && *sz.Face.Index < len(imageMeta.Faces) {
		faceBounds = imageMeta.Faces[*sz.Face.Index].Box
	} else {
		if sz.Face.Largest != nil {
			area := 0.0
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea > area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else if sz.Face.Smallest != nil {
			area := math.MaxFloat64
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea < area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else {
			faceBounds = vision.CalcFacesBounds(imageMeta.Faces)
		}
	}

	if sz.Face.Focus != nil && *sz.Face.Focus {
		return sz.cropFocus(faceBounds.Left+(faceBounds.Width/2.0), faceBounds.Top+(faceBounds.Height/2.0), cW, cH, params, sourceImage, sz.Face.Zoom)
	}

	return sz.cropBounds(
		cW,
		cH,
		faceBounds,
		sz.Zoom,
		sz.Face.Zoom,
		utils.IfNil(sz.Face.Padding, 0),
		utils.IfNil(sz.Face.HGravity, "center"),
		utils.IfNil(sz.Face.VGravity, "top"),
		params,
		sourceImage,
	)
}

func (sz *SizingOptions) cropPerson(cW *int, cH *int, imageMeta *vision.Metadata, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	if sz.Person == nil {
		return sz.cropFill(cW, cH, params, sourceImage)
	}

	var personBounds geometry.Box
	if sz.Person.Index != nil && *sz.Person.Index < len(imageMeta.People) {
		if imageMeta.People[*sz.Person.Index].Box == nil {
			return sz.cropFill(cW, cH, params, sourceImage)
		} else {
			personBounds = *imageMeta.People[*sz.Person.Index].Box
		}
	} else {
		if sz.Person.Largest != nil {
			area := 0.0
			for _, person := range imageMeta.People {
				if person.Box == nil {
					continue
				}

				personArea := person.Box.Width * person.Box.Height
				if personArea > area {
					personBounds = *person.Box
					area = personArea
				}
			}
		} else if sz.Person.Smallest != nil {
			area := math.MaxFloat64
			for _, person := range imageMeta.People {
				if person.Box == nil {
					continue
				}

				personArea := person.Box.Width * person.Box.Height
				if personArea < area {
					personBounds = *person.Box
					area = personArea
				}
			}
		} else {
			personBounds = vision.CalcLabelsBounds(imageMeta.People)
		}
	}

	if utils.IfNil(sz.Person.Focus, false) {
		return sz.cropFocus(personBounds.Left+(personBounds.Width/2.0), personBounds.Top+(personBounds.Height/2.0), cW, cH, params, sourceImage, sz.Person.Zoom)
	}

	return sz.cropBounds(
		cW,
		cH,
		personBounds,
		sz.Zoom,
		sz.Person.Zoom,
		utils.IfNil(sz.Face.Padding, 0),
		utils.IfNil(sz.Face.HGravity, "center"),
		utils.IfNil(sz.Face.VGravity, "center"),
		params,
		sourceImage,
	)
}

func (sz *SizingOptions) cropFill(cW *int, cH *int, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sw := sourceImage.Width()
	sh := sourceImage.Height()

	var targetWidth int
	var targetHeight int

	if cW == nil {
		targetWidth = 10000
	} else {
		targetWidth = *cW
	}

	if cH == nil {
		targetHeight = 10000
	} else {
		targetHeight = *cH
	}

	cropSize := geometry.SizeToFitSize(targetWidth, targetHeight, sw, sh)
	if sz.Zoom != nil {
		cropSize.Width = int(math.Round(float64(cropSize.Width) * (1 / *sz.Zoom)))
		cropSize.Height = int(math.Round(float64(cropSize.Height) * (1 / *sz.Zoom)))
	}

	cropX := utils.Min(sw, utils.Max(0, (sw/2)-(cropSize.Width/2)))
	cropY := utils.Min(sh, utils.Max(0, (sh/2)-(cropSize.Height/2)))

	if utils.IfNil(sz.HGravity, "center") == "left" {
		cropX = 0
	} else if utils.IfNil(sz.HGravity, "center") == "right" {
		cropX = sw - cropSize.Width
	}

	if utils.IfNil(sz.VGravity, "center") == "top" {
		cropY = 0
	} else if utils.IfNil(sz.VGravity, "center") == "bottom" {
		cropY = sh - cropSize.Height
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

	if (cropSize.Width != targetWidth) || (cropSize.Height != targetHeight) {
		err = sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}

func (sz *SizingOptions) cropFit(cW *int, cH *int, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	sourceImage, err := sz.resize(params, sourceImage)
	if err != nil {
		return sourceImage, err
	}

	if cW != nil && cH != nil {
		transparent := params.Export != nil && params.Export.Format != nil && (*params.Export.Format == "png" || *params.Export.Format == "webp")
		backgroundColor, bgColorErr := ParseHexColor(utils.IfNil(params.Background.Color, "00000000"))
		if bgColorErr != nil {
			backgroundColor = ColorRGBA{R: 0, G: 0, B: 0, A: 0}
		}

		if transparent {
			if !sourceImage.HasAlpha() {
				err = sourceImage.AddAlpha()
				if err != nil {
					return sourceImage, err
				}
			}

			err = sourceImage.EmbedBackgroundRGBA(
				(*cW/2)-(sourceImage.Width()/2),
				(*cH/2)-(sourceImage.Height()/2),
				*cW,
				*cH,
				&vips.ColorRGBA{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B, A: backgroundColor.A},
			)

			if err != nil {
				return sourceImage, err
			}
		} else {
			err = sourceImage.EmbedBackground(
				(*cW/2)-(sourceImage.Width()/2),
				(*cH/2)-(sourceImage.Height()/2),
				*cW,
				*cH,
				&vips.Color{R: backgroundColor.R, G: backgroundColor.G, B: backgroundColor.B},
			)

			if err != nil {
				return sourceImage, err
			}
		}
	}

	return sourceImage, nil
}

func (sz *SizingOptions) cropFocus(focusX float64, focusY float64, cW *int, cH *int, params *ImageParams, sourceImage *vips.ImageRef, zoom *float64) (*vips.ImageRef, error) {
	if params == nil {
		return sz.cropFill(cW, cH, params, sourceImage)
	}

	sw := sourceImage.Width()
	sh := sourceImage.Height()

	fpx := int(math.Round(focusX * float64(sw)))
	fpy := int(math.Round(focusY * float64(sh)))

	var newCW int
	var newCH int
	if *cW < *cH {
		newCH = utils.Max(*cH, utils.Min(fpy, sw-fpy))
		newCW = int(math.Round(float64(newCH) * (float64(*cW) / float64((*cH)))))
	} else if *cW > *cH {
		newCW = utils.Max(*cW, utils.Min(fpx, sw-fpx))
		newCH = int(math.Round(float64(newCW) * (float64(*cH) / float64(*cW))))
	} else {
		newCW = utils.Max(*cW, utils.Min(fpx, sw-fpx, fpy, sh-fpy))
		newCH = newCW
	}

	if zoom != nil && *zoom > 0 {
		newCW = int(math.Round(float64(newCW) * (1.0 / *zoom)))
		newCH = int(math.Round(float64(newCH) * (1.0 / *zoom)))
	}

	var cropSize geometry.Size
	if newCW*2 > sw || newCH*2 > sh {
		cropSize = geometry.SizeToFitSize(newCW*2, newCH*2, sw, sh)
	} else {
		cropSize = geometry.Size{Width: newCW * 2, Height: newCH * 2}
	}

	if sz.Zoom != nil {
		cropSize.Width = int(math.Round(float64(cropSize.Width) * (1 / *sz.Zoom)))
		cropSize.Height = int(math.Round(float64(cropSize.Height) * (1 / *sz.Zoom)))
	}

	cropX := utils.Min(sw-cropSize.Width, utils.Max(0, fpx-(cropSize.Width/2)))
	cropY := utils.Min(sh-cropSize.Height, utils.Max(0, fpy-(cropSize.Height/2)))

	if utils.IfNil(sz.HGravity, "center") == "left" {
		cropX = 0
	} else if utils.IfNil(sz.HGravity, "center") == "right" {
		cropX = sw - cropSize.Width
	}

	if utils.IfNil(sz.VGravity, "center") == "top" {
		cropY = 0
	} else if utils.IfNil(sz.VGravity, "center") == "bottom" {
		cropY = sh - cropSize.Height
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

	if (cropSize.Width != *cW) || (cropSize.Height != *cH) {
		err = sourceImage.ThumbnailWithSize(*cW, *cH, vips.InterestingNone, vips.SizeDown)
		if err != nil {
			log.Fatal("Faces Thumbnail With Size Error:", err)
		}
	}

	return sourceImage, nil
}

func (sz *SizingOptions) resize(params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var targetWidth int
	var targetHeight int

	if sz.Width == nil {
		targetWidth = 10000
	} else {
		targetWidth = *sz.Width
	}

	if sz.Height == nil {
		targetHeight = 10000
	} else {
		targetHeight = *sz.Height
	}

	err := sourceImage.ThumbnailWithSize(targetWidth, targetHeight, vips.InterestingNone, vips.SizeBoth)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}

func (sz *SizingOptions) cropSmart(cW *int, cH *int, params *ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var interesting = vips.InterestingAttention
	if sz.Interesting != nil {
		interesting = *sz.Interesting
	}

	if sz.Zoom != nil {
		*cW = int(math.Round(float64(*cW) * (1 / *sz.Zoom)))
		*cH = int(math.Round(float64(*cH) * (1 / *sz.Zoom)))
	}

	err := sourceImage.SmartCrop(*cW, *cH, interesting)
	if err != nil {
		return nil, err
	}

	return sourceImage, nil
}
