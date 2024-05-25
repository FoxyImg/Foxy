package images

import (
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

func cropPerson(cW *int, cH *int, imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) (*vips.ImageRef, error) {
	var personBounds geometry.Box
	if params.Person.Index != nil && *params.Person.Index < len(imageMeta.People) {
		if imageMeta.People[*params.Person.Index].Box == nil {
			return cropFill(cW, cH, params, sourceImage)
		} else {
			personBounds = *imageMeta.People[*params.Person.Index].Box
		}
	} else {
		if params.Person.Largest != nil {
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
		} else if params.Person.Smallest != nil {
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
			personBounds = metadata.CalcLabelsBounds(imageMeta.People)
		}
	}

	return cropBounds(
		cW,
		cH,
		personBounds,
		params.Zoom,
		params.Person.Zoom,
		params.Person.Padding,
		params.Person.HGravity,
		params.Person.VGravity,
		params,
		sourceImage,
	)
}
