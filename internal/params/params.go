package params

import (
	"foxy/internal/utils"
	"slices"
	"strings"
)

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

	Rotation *RotationParams `json:"rotation,omitempty"`
	Size     *SizingOptions  `json:"size,omitempty"`

	Background *BackgroundOptions `json:"background,omitempty"`

	Padding *PadOptions    `json:"padding,omitempty"`
	Border  *BorderOptions `json:"border,omitempty"`

	Redact      *RedactOptions     `json:"redact,omitempty"`
	Stylize     *StylizeParams     `json:"stylize,omitempty"`
	GradientMap *GradientMapParams `json:"gradientMap,omitempty"`
	Adjustments *AdjustmentsParams `json:"adjustments,omitempty"`

	Watermark *WatermarkParams `json:"watermark,omitempty"`

	Export *ExportOptions `json:"export,omitempty"`

	FlipH *bool `json:"flipH,omitempty"`
	FlipV *bool `json:"flipV,omitempty"`
}

func NewImageParams() *ImageParams {
	return &ImageParams{
		Debug:       &DebugOptions{},
		Background:  &BackgroundOptions{},
		Rotation:    &RotationParams{},
		Size:        &SizingOptions{},
		Border:      &BorderOptions{},
		Padding:     &PadOptions{},
		Redact:      &RedactOptions{},
		Stylize:     &StylizeParams{},
		Watermark:   &WatermarkParams{},
		GradientMap: &GradientMapParams{},
		Adjustments: &AdjustmentsParams{},
		Export: &ExportOptions{
			Format:  utils.Ptr("webp"),
			Quality: utils.Ptr(85),
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
		} else if slices.Contains(result.Rotation.Params(), split[0]) {
			_ = result.Rotation.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Adjustments.Params(), split[0]) {
			_ = result.Adjustments.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.GradientMap.Params(), split[0]) {
			_ = result.GradientMap.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Stylize.Params(), split[0]) {
			_ = result.Stylize.ParseParams(split[0], split[1:])
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
		} else if slices.Contains(result.Background.Params(), split[0]) {
			_ = result.Background.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Export.Params(), split[0]) {
			_ = result.Export.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Watermark.Params(), split[0]) {
			_ = result.Watermark.ParseParams(split[0], split[1:])
		} else if split[0] == "meta" {
			result.MetaOnly = true
			result.NeedsVision = true
		}

	}

	return &result, nil
}
