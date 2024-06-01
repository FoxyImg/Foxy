package images

import (
	"fmt"
	"foxy/internal/utils"
	"foxy/internal/vision"
	"github.com/davidbyttow/govips/v2/vips"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
)

type RedactOptions struct {
	Faces        *[]int  `json:"faces,omitempty"`
	People       *[]int  `json:"people,omitempty"`
	Regions      *[]Rect `json:"regions,omitempty"`
	Blur         *int    `json:"blur,omitempty"`
	BlurMask     *int    `json:"blurMask,omitempty"`
	ExpandMask   *int    `json:"expandMask,omitempty"`
	UseColor     *bool   `json:"useColor,omitempty"`
	Color        *string `json:"color,omitempty"`
	Pixelate     *int    `json:"pixelate,omitempty"`
	PixelateMask *int    `json:"pixelateMask,omitempty"`
}

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

func (redact *RedactOptions) Params() []string {
	return []string{"redact"}
}

func (redact *RedactOptions) ParseParams(param string, options []string) (needsVision bool) {
	if len(options) < 2 {
		return
	}

	needsVision = true

	if options[0] == "faces" && len(options) == 2 {
		faces := strings.Split(options[1], ",")

		if redact.Faces == nil {
			redact.Faces = utils.Ptr([]int{})
		}

		if slices.Contains(faces, "all") {
			*redact.Faces = append(*redact.Faces, -1)
		} else {
			for _, face := range faces {
				f, err := strconv.Atoi(face)
				if err != nil {
					continue
				}

				*redact.Faces = append(*redact.Faces, f)
			}
		}
	} else if options[0] == "people" && len(options) == 2 {
		people := strings.Split(options[1], ",")

		if redact.People == nil {
			redact.People = utils.Ptr([]int{})
		}

		if slices.Contains(people, "all") {
			*redact.People = append(*redact.People, -1)
		} else {
			for _, person := range people {
				p, err := strconv.Atoi(person)
				if err != nil {
					continue
				}

				*redact.People = append(*redact.People, p)
			}
		}
	} else if options[0] == "region" && len(options) == 2 {
		regionParts := strings.Split(options[1], ",")

		if len(regionParts) == 4 {
			if redact.Regions == nil {
				redact.Regions = utils.Ptr([]Rect{})
			}

			l, _ := strconv.ParseFloat(regionParts[0], 64)
			t, _ := strconv.ParseFloat(regionParts[1], 64)
			w, _ := strconv.ParseFloat(regionParts[2], 64)
			h, _ := strconv.ParseFloat(regionParts[3], 64)

			r := Rect{
				Left:   l,
				Top:    t,
				Width:  w,
				Height: h,
			}

			*redact.Regions = append(*redact.Regions, r)
		}
	} else if options[0] == "blur" && len(options) == 2 {
		b, _ := strconv.Atoi(options[1])
		redact.Blur = utils.Ptr(b)
	} else if options[0] == "color" && len(options) == 2 {
		redact.UseColor = utils.Ptr(true)
		redact.Color = &options[1]
	} else if options[0] == "pixelate" {
		p, _ := strconv.Atoi(options[1])
		redact.Pixelate = utils.Ptr(p)
	} else if options[0] == "mask" && len(options) == 3 {
		if options[1] == "blur" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				redact.BlurMask = utils.Ptr(b)
			}
		} else if options[1] == "expand" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				redact.ExpandMask = utils.Ptr(b)
			}
		} else if options[1] == "pixelate" {
			b, err := strconv.Atoi(options[2])
			if err == nil {
				redact.PixelateMask = utils.Ptr(b)
			}
		}
	}

	return
}

func (redact *RedactOptions) Process(sourceImage *vips.ImageRef, params *ImageParams, imageMeta *vision.Metadata) (*vips.ImageRef, error) {
	if utils.IfNil(redact.Blur, 0) == 0 && utils.IfNil(redact.Pixelate, 0) == 0 && utils.IfNil(redact.UseColor, false) == false {
		return sourceImage, nil
	}

	if (redact.Faces == nil || len(*redact.Faces) == 0) && (redact.People == nil || len(*redact.People) == 0) && (redact.Regions == nil || len(*redact.Regions) == 0) {
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
	if utils.IfNil(redact.UseColor, false) && redact.Color != nil {
		color = "#" + *redact.Color
	} else {
		color = "black"
	}

	if redact.Faces != nil && slices.Contains(*redact.Faces, -1) {
		for _, face := range imageMeta.Faces {
			rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", utils.IfNil(redact.ExpandMask, 0), color, false)
			overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", utils.IfNil(redact.ExpandMask, 0), color, true)
		}
	} else if redact.Faces != nil {
		for _, faceIdx := range *redact.Faces {
			if faceIdx > 0 && faceIdx < len(imageMeta.Faces) {
				face := imageMeta.Faces[faceIdx]
				rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", utils.IfNil(redact.ExpandMask, 0), color, false)
				overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), face.Box.Left, face.Box.Top, face.Box.Width, face.Box.Height, "100%", "100%", utils.IfNil(redact.ExpandMask, 0), color, true)
			}
		}
	}

	if redact.People != nil && slices.Contains(*redact.People, -1) {
		for _, person := range imageMeta.People {
			rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", utils.IfNil(redact.ExpandMask, 0), color, false)
			overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", utils.IfNil(redact.ExpandMask, 0), color, true)
		}
	} else if redact.People != nil {
		for _, personIdx := range *redact.People {
			if personIdx > 0 && personIdx < len(imageMeta.People) {
				person := imageMeta.Faces[personIdx]
				rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", utils.IfNil(redact.ExpandMask, 0), color, false)
				overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), person.Box.Left, person.Box.Top, person.Box.Width, person.Box.Height, "0", "0", utils.IfNil(redact.ExpandMask, 0), color, true)
			}
		}
	}

	if redact.Regions != nil {
		for _, region := range *redact.Regions {
			rects = rects + svgRect(redactedParts.Width(), redactedParts.Height(), region.Left, region.Top, region.Width, region.Height, "0", "0", 0, color, false)
			overlayRects = overlayRects + svgRect(redactedParts.Width(), redactedParts.Height(), region.Left, region.Top, region.Width, region.Height, "0", "0", 0, color, true)
		}
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

	if utils.IfNil(redact.UseColor, false) && utils.IfNil(redact.Blur, 0) == 0 && utils.IfNil(redact.Pixelate, 0) == 0 {
		colorMask, err := svg.Copy()
		if err != nil {
			return sourceImage, err
		}

		if redact.BlurMask != nil && *redact.BlurMask > 0 {
			_ = colorMask.GaussianBlur(float64(*redact.BlurMask))
		}

		if redact.PixelateMask != nil && *redact.PixelateMask > 0 {
			_ = colorMask.Resize(1.0/float64(*redact.PixelateMask), vips.KernelLanczos3)
			_ = colorMask.Resize(float64(*redact.PixelateMask), vips.KernelNearest)
		}

		if redactedScale != 1.0 {
			_ = colorMask.Resize(1.0/redactedScale, vips.KernelLanczos3)
		}

		_ = sourceImage.Composite(colorMask, vips.BlendModeOver, 0, 0)
	} else if utils.IfNil(redact.Blur, 0) > 0 || utils.IfNil(redact.Blur, 0) > 0 {
		if utils.IfNil(redact.UseColor, false) {
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

		if redact.Blur != nil && *redact.Blur > 0 {
			_ = redactedParts.GaussianBlur(float64(*redact.Blur))
		}

		if redact.Pixelate != nil && *redact.Pixelate > 0 {
			_ = redactedParts.Resize(1.0/float64(*redact.Pixelate), vips.KernelLanczos3)
			_ = redactedParts.Resize(float64(*redact.Pixelate), vips.KernelNearest)
		}

		if redact.BlurMask != nil && *redact.BlurMask > 0 {
			_ = alpha.GaussianBlur(float64(*redact.BlurMask))
		}

		if redact.PixelateMask != nil && *redact.PixelateMask > 0 {
			_ = alpha.Resize(1.0/float64(*redact.PixelateMask), vips.KernelLanczos3)
			_ = alpha.Resize(float64(*redact.PixelateMask), vips.KernelNearest)
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
