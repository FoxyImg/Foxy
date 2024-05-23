package images

import (
	"errors"
	"foxy/internal/metadata"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func Crop(imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	if params.Width != nil || params.Height != nil {
		cW := params.Width
		cH := params.Height

		if params.AspectRatio != nil {
			if cW != nil {
				nh := int(math.Floor(float64(*cW) / *params.AspectRatio))
				cH = &nh
			} else if cH != nil {
				nw := int(math.Floor(float64(*cH) * *params.AspectRatio))
				cW = &nw
			}
		}

		if params.Crop != nil {
			for _, cropMode := range *params.Crop {
				if cropMode == "face" && imageMeta != nil && len(imageMeta.Faces) > 0 && cW != nil && cH != nil {
					croppedImage, err := cropFace(cW, cH, imageMeta, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "person" && imageMeta != nil && len(imageMeta.People) > 0 && cW != nil && cH != nil {
					croppedImage, err := cropPerson(cW, cH, imageMeta, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "fit" && cW != nil && cH != nil {
					croppedImage, err := cropFit(cW, cH, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "smart" && cW != nil && cH != nil {
					croppedImage, err := cropSmart(cW, cH, params, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				} else if cropMode == "fill" && cW != nil && cH != nil {
					croppedImage, err := cropFill(cW, cH, sourceImage)
					if err != nil {
						return nil, err
					}

					return croppedImage, nil
				}
			}

			if cW != nil && cH != nil {
				croppedImage, err := cropFill(cW, cH, sourceImage)
				if err != nil {
					return nil, err
				}

				return croppedImage, nil
			} else if cW != nil || cH != nil {
				resizedImage, err := resize(params, sourceImage)
				if err != nil {
					return nil, err
				}

				return resizedImage, nil
			} else {
				return nil, errors.New("unknown crop mode")
			}
		} else {
			resizedImage, err := resize(params, sourceImage)
			if err != nil {
				return nil, err
			}

			return resizedImage, nil
		}
	}

	return sourceImage, nil
}
