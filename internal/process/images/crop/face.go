package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func cropFace(cW *int, cH *int, imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var faceBounds geometry.Box
	if params.Face.Index != nil && *params.Face.Index < len(imageMeta.Faces) {
		faceBounds = imageMeta.Faces[*params.Face.Index].Box
	} else {
		if params.Face.Largest != nil {
			area := 0.0
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea > area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else if params.Face.Smallest != nil {
			area := math.MaxFloat64
			for _, face := range imageMeta.Faces {
				faceArea := face.Box.Width * face.Box.Height
				if faceArea < area {
					faceBounds = face.Box
					area = faceArea
				}
			}
		} else {
			faceBounds = metadata.CalcFacesBounds(imageMeta.Faces)
		}
	}

	if params.Face.Focus {
		return cropFocus(faceBounds.Left+(faceBounds.Width/2.0), faceBounds.Top+(faceBounds.Height/2.0), cW, cH, params, sourceImage, params.Face.Zoom)
	}

	return cropBounds(
		cW,
		cH,
		faceBounds,
		params.Zoom,
		params.Face.Zoom,
		params.Face.Padding,
		params.Face.HGravity,
		params.Face.VGravity,
		params,
		sourceImage,
	)
}
