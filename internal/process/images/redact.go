package images

import (
	"fmt"
	"foxy/internal/metadata"
	"foxy/internal/utils"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"slices"
)

func svgRect(sw int, sh int, l float64, t float64, w float64, h float64, rx string, ry string, expand int, color string, useAlpha bool) string {
	newColor := color
	if len(newColor) > 7 && !useAlpha {
		newColor = newColor[0:7]
	}

	if expand > 0 {
		expandX := w * (float64(expand) / 100.0)
		expandY := h * (float64(expand) / 100.0)

		l = l - (expandX / 2.0)
		t = t - (expandY / 2.0)
		w = w + expandX
		h = h + expandY
	}

	rl := int(math.Round(l * float64(sw)))
	rt := int(math.Round(t * float64(sh)))
	rw := int(math.Round(w * float64(sw)))
	rh := int(math.Round(h * float64(sh)))

	return fmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" rx=\"%s\" ry=\"%s\" style=\"fill:%s;\" />",
		rl, rt, rw, rh, rx, ry, newColor,
	)
}

func Redact(sourceImage *vips.ImageRef, params *metadata.ImageParams, imageMeta *metadata.Metadata) (*vips.ImageRef, error) {
	if params.Redact.Blur == 0 && params.Redact.Pixelate == 0 && params.Redact.UseColor == false {
		return sourceImage, nil
	}

	if len(params.Redact.Faces) == 0 && len(params.Redact.People) == 0 && len(params.Redact.Regions) == 0 {
		return sourceImage, nil
	}

	var redactedScale = 1.0
	redactedParts, err := sourceImage.Copy()
	if err != nil {
		return sourceImage, err
	}

	if utils.Max(redactedParts.Width(), redactedParts.Height()) > 1920 {
		redactedScale = 1920.0 / float64(utils.Max(redactedParts.Width(), redactedParts.Height()))
		err = redactedParts.Resize(redactedScale, vips.KernelLanczos3)

		newImageBuffer, _, err := redactedParts.ExportNative()
		if err != nil {
			return sourceImage, err
		}

		redactedParts, err = vips.NewImageFromBuffer(newImageBuffer)
		if err != nil {
			return sourceImage, err
		}
	}

	var overlayRects = ""
	var rects = ""

	var color string
	if params.Redact.UseColor && params.Redact.Color != nil {
		color = "#" + *params.Redact.Color
	} else {
		color = "black"
	}

	if slices.Contains(params.Redact.Faces, -1) {
		for _, face := range imageMeta.Faces {
			rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", params.Redact.ExpandMask, color, false)
			overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", params.Redact.ExpandMask, color, true)
		}
	} else {
		for _, faceIdx := range params.Redact.Faces {
			if faceIdx > 0 && faceIdx < len(imageMeta.Faces) {
				face := imageMeta.Faces[faceIdx]
				rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", params.Redact.ExpandMask, color, false)
				overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", params.Redact.ExpandMask, color, true)
			}
		}
	}

	if slices.Contains(params.Redact.People, -1) {
		for _, person := range imageMeta.People {
			rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", params.Redact.ExpandMask, color, false)
			overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", params.Redact.ExpandMask, color, true)
		}
	} else {
		for _, personIdx := range params.Redact.People {
			if personIdx > 0 && personIdx < len(imageMeta.People) {
				person := imageMeta.Faces[personIdx]
				rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", params.Redact.ExpandMask, color, false)
				overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", params.Redact.ExpandMask, color, true)
			}
		}
	}

	for _, region := range params.Redact.Regions {
		rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), region.Left, region.Top, region.Width, region.Height, "0", "0", 0, color, false)
		overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), region.Left, region.Top, region.Width, region.Height, "0", "0", 0, color, true)
	}

	if rects == "" {
		return sourceImage, nil
	}

	svgString := fmt.Sprintf("<svg width=\"%d\" height=\"%d\" viewport=\"0 0 %d %d\" xmlns=\"http://www.w3.org/2000/svg\">%s</svg>",
		redactedParts.Width(),
		redactedParts.Height(),
		redactedParts.Width(),
		redactedParts.Height(),
		rects,
	)

	overlaySvgString := fmt.Sprintf("<svg width=\"%d\" height=\"%d\" viewport=\"0 0 %d %d\" xmlns=\"http://www.w3.org/2000/svg\">%s</svg>",
		redactedParts.Width(),
		redactedParts.Height(),
		redactedParts.Width(),
		redactedParts.Height(),
		overlayRects,
	)

	log.Println("Redact SVG:", svgString)
	log.Println("Redact SVG:", overlaySvgString)

	svg, err := vips.LoadImageFromBuffer([]byte(svgString), vips.NewImportParams())
	if err != nil {
		return sourceImage, err
	}

	if params.Redact.UseColor && params.Redact.Blur == 0 && params.Redact.Pixelate == 0 {
		colorMask, err := svg.Copy()
		if err != nil {
			return sourceImage, err
		}

		if params.Redact.BlurMask > 0 {
			_ = colorMask.GaussianBlur(float64(params.Redact.BlurMask))
		}

		if params.Redact.PixelateMask > 0 {
			_ = colorMask.Resize(1.0/float64(params.Redact.PixelateMask), vips.KernelLanczos3)
			_ = colorMask.Resize(float64(params.Redact.PixelateMask), vips.KernelNearest)
		}

		if redactedScale != 1.0 {
			_ = colorMask.Resize(1.0/redactedScale, vips.KernelLanczos3)
		}

		_ = sourceImage.Composite(colorMask, vips.BlendModeOver, 0, 0)
	} else if params.Redact.Blur > 0 || params.Redact.Pixelate > 0 {
		if params.Redact.UseColor {
			overlaySvg, err := vips.LoadImageFromBuffer([]byte(overlaySvgString), vips.NewImportParams())
			if err != nil {
				return sourceImage, err
			}

			_ = redactedParts.Composite(overlaySvg, vips.BlendModeOver, 0, 0)
		}

		alpha, err := svg.ExtractBandToImage(3, 1)
		if err != nil {
			log.Println("Extract Band To Image Error:", err)
			return sourceImage, err
		}

		if redactedParts.HasAlpha() {
			err = redactedParts.ExtractBand(0, redactedParts.Bands()-1)
			if err != nil {
				log.Println("Extract Band Error:", err)
				return sourceImage, err
			}
		}

		if params.Redact.Blur > 0 {
			_ = redactedParts.GaussianBlur(float64(params.Redact.Blur))
		}

		if params.Redact.Pixelate > 0 {
			_ = redactedParts.Resize(1.0/float64(params.Redact.Pixelate), vips.KernelLanczos3)
			_ = redactedParts.Resize(float64(params.Redact.Pixelate), vips.KernelNearest)
		}

		if params.Redact.BlurMask > 0 {
			_ = alpha.GaussianBlur(float64(params.Redact.BlurMask))
		}

		if params.Redact.PixelateMask > 0 {
			_ = alpha.Resize(1.0/float64(params.Redact.PixelateMask), vips.KernelLanczos3)
			_ = alpha.Resize(float64(params.Redact.PixelateMask), vips.KernelNearest)
		}

		err = redactedParts.BandJoin(alpha)
		if err != nil {
			log.Println("Band Join Error:", err)
			return sourceImage, err
		}

		if redactedScale != 1.0 {
			_ = redactedParts.Resize(1.0/redactedScale, vips.KernelLanczos3)
		}

		_ = sourceImage.Composite(redactedParts, vips.BlendModeOver, 0, 0)
	}

	return sourceImage, nil
}
