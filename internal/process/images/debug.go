package images

import (
	"fmt"
	"foxy/internal/geometry"
	"foxy/internal/metadata"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"strconv"
)

func drawBounds(label string, image *vips.ImageRef, box geometry.Box, color vips.ColorRGBA) {
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

func drawDebugBounds(imageMeta *metadata.Metadata, params *metadata.ImageParams, sourceImage *vips.ImageRef) {
	if params.Debug.OtherLabels {
		otherLabel := metadata.FilterNonPersonLabels(imageMeta.Labels)
		for _, label := range otherLabel {
			drawBounds(label.Name, sourceImage, *label.Box, vips.ColorRGBA{R: 255, G: 0, B: 255, A: 255})
		}
	}

	if params.Debug.AllPeople && len(imageMeta.People) > 0 {
		drawBounds("all people", sourceImage, metadata.CalcLabelsBounds(imageMeta.People), vips.ColorRGBA{R: 0, G: 255, B: 255, A: 255})
	}

	if params.Debug.People && len(imageMeta.People) > 0 {
		for idx, label := range imageMeta.People {
			if label.Box != nil {
				log.Println("Person #", idx, label.Name)
				drawBounds("#"+strconv.Itoa(idx)+" - "+label.Name, sourceImage, *label.Box, vips.ColorRGBA{R: 0, G: 255, B: 0, A: 255})
			}
		}
	}

	if params.Debug.AllFaces && len(imageMeta.Faces) > 0 {
		drawBounds("all faces", sourceImage, metadata.CalcFacesBounds(imageMeta.Faces), vips.ColorRGBA{R: 255, G: 255, B: 0, A: 255})
	}

	if params.Debug.Faces && len(imageMeta.Faces) > 0 {
		for idx, face := range imageMeta.Faces {
			drawBounds("face #"+strconv.Itoa(idx), sourceImage, face.Box, vips.ColorRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
}
