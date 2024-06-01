package images

import (
	"slices"
	"strconv"
	"strings"
)

type ImageExportParams struct {
	Format          string `json:"format"`
	Quality         int    `json:"quality"`
	ReductionEffort *int   `json:"reductionEffort"`
	Lossless        *bool  `json:"lossless"`
	NearLossless    *bool  `json:"nearLossless"`
}

type ColorRGBA struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Rect struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ImageParams struct {
	Debug *DebugOptions `json:"-"`

	MetaOnly    bool `json:"-"`
	NeedsVision bool `json:"vision"`

	Size *SizingOptions `json:"size,omitempty"`

	BackgroundColor string `json:"bgColor"`

	Padding *PadOptions    `json:"padding,omitempty"`
	Border  *BorderOptions `json:"border,omitempty"`

	Redact  *RedactOptions `json:"redact,omitempty"`
	Stylize *StylizeParams `json:"stylize,omitempty"`

	Rotate *int  `json:"rotate,omitempty"`
	FlipH  *bool `json:"flipH,omitempty"`
	FlipV  *bool `json:"flipV,omitempty"`

	Brightness *float64 `json:"brightness,omitempty"`
	Saturation *float64 `json:"saturation,omitempty"`
	Hue        *float64 `json:"hue,omitempty"`

	ExportParams ImageExportParams `json:"export"`
}

func NewImageParams() *ImageParams {
	return &ImageParams{
		Debug:           &DebugOptions{},
		BackgroundColor: "#00000000",
		Size:            &SizingOptions{},
		Border:          &BorderOptions{},
		Padding:         &PadOptions{},
		Redact:          &RedactOptions{},
		Stylize:         &StylizeParams{},
		ExportParams: ImageExportParams{
			Format:  "webp",
			Quality: 85,
		},
	}
}

func BuildParams(pathParts []string) (*ImageParams, error) {
	result := *NewImageParams()

	for _, part := range pathParts {
		split := strings.Split(part, ":")

		if slices.Contains(result.Size.Params(), split[0]) {
			nv := result.Size.ParseParams(split[0], split[1:])
			result.NeedsVision = result.NeedsVision || nv
		} else if slices.Contains(result.Stylize.Params(), split[0]) {
			nv := result.Stylize.ParseParams(split[0], split[1:])
			result.NeedsVision = result.NeedsVision || nv
		} else if slices.Contains(result.Border.Params(), split[0]) {
			_ = result.Border.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Padding.Params(), split[0]) {
			_ = result.Padding.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Redact.Params(), split[0]) {
			nv := result.Redact.ParseParams(split[0], split[1:])
			result.NeedsVision = result.NeedsVision || nv
		} else if slices.Contains(result.Debug.Params(), split[0]) {
			nv := result.Debug.ParseParams(split[0], split[1:])
			result.NeedsVision = result.NeedsVision || nv
		} else {
			switch split[0] {
			case "bg":
				if len(split) != 2 {
					continue
				}

				result.BackgroundColor = split[1]
			// File Format Related
			case "fmt":
				if len(split) != 2 {
					continue
				}

				result.ExportParams.Format = split[1]
			case "q":
				if len(split) != 2 {
					continue
				}

				q, err := strconv.Atoi(split[1])
				if err != nil {
					continue
				}

				result.ExportParams.Quality = q
			case "nloss":
				if len(split) != 2 {
					continue
				}

				nloss := split[1] == "1"

				result.ExportParams.NearLossless = &nloss
			case "lossless":
				if len(split) != 2 {
					continue
				}

				loss := split[1] == "1"

				result.ExportParams.Lossless = &loss
			case "reduction":
				if len(split) != 2 {
					continue
				}

				e, err := strconv.Atoi(split[1])
				if err != nil {
					continue
				}

				result.ExportParams.ReductionEffort = &e
			// Metadata
			case "meta":
				result.MetaOnly = true
				result.NeedsVision = true
			default:
			}
		}

	}

	return &result, nil
}
