package params

import (
	"foxy/internal/utils"
	"net/url"
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
	Debug *DebugOptions `json:"debug,omitempty"`

	MetaOnly    bool `json:"-"`
	NeedsVision bool `json:"vision"`

	BackgroundRemoval *BackgroundRemovalParams `json:"backgroundRemoval,omitempty"`

	Levels *LevelsParams `json:"levels,omitempty"`

	SourceCrop *SourceCropParams `json:"sourceCrop,omitempty"`

	Rotation *RotationParams `json:"rotation,omitempty"`
	Sizing   *SizingOptions  `json:"sizing,omitempty"`

	Background *BackgroundOptions `json:"background,omitempty"`

	Padding *PadOptions    `json:"padding,omitempty"`
	Border  *BorderOptions `json:"border,omitempty"`

	Redact      *RedactOptions     `json:"redact,omitempty"`
	Stylize     *StylizeParams     `json:"stylize,omitempty"`
	GradientMap *GradientMapParams `json:"gradientMap,omitempty"`
	Adjustments *AdjustmentsParams `json:"adjustments,omitempty"`

	Mask *MaskParams `json:"mask,omitempty"`

	Export *ExportOptions `json:"export,omitempty"`

	Overlays Overlays `json:"overlays,omitempty"`

	Video *VideoParams `json:"video,omitempty"`
}

func NewImageParams() *ImageParams {
	return &ImageParams{
		BackgroundRemoval: &BackgroundRemovalParams{},
		SourceCrop:        &SourceCropParams{},
		Debug:             &DebugOptions{},
		Background:        &BackgroundOptions{},
		Levels:            &LevelsParams{},
		Rotation:          &RotationParams{},
		Sizing:            &SizingOptions{},
		Border:            &BorderOptions{},
		Padding:           &PadOptions{},
		Redact:            &RedactOptions{},
		Stylize:           &StylizeParams{},
		GradientMap:       &GradientMapParams{},
		Adjustments:       &AdjustmentsParams{},
		Mask:              &MaskParams{},
		Overlays:          make(Overlays),

		Export: &ExportOptions{
			Format:  utils.Ptr("webp"),
			Quality: utils.Ptr(85),
		},

		Video: &VideoParams{
			Type:      utils.Ptr("frame"),
			FrameType: utils.Ptr("rel"),
			Time:      utils.Ptr(0.5),
		},
	}
}

func BuildParamsFromQuery(values url.Values) (*ImageParams, error) {
	var pathParts []string
	for key, value := range values {
		if key == "_" {
			continue
		}

		if key == "s" {
			continue
		}

		if key == "showpreset" {
			continue
		}

		if value == nil || len(value) == 0 || value[0] == "" {
			pathParts = append(pathParts, strings.ReplaceAll(key, "-", ":"))
		} else {
			pathParts = append(pathParts, strings.ReplaceAll(key, "-", ":")+":"+strings.Join(value, ","))
		}
	}

	return BuildParams(pathParts)
}

func BuildParams(pathParts []string) (*ImageParams, error) {
	result := *NewImageParams()

	for _, part := range pathParts {
		split := strings.Split(part, ":")

		if slices.Contains(result.Sizing.Params(), split[0]) {
			nv := result.Sizing.ParseParams(split[0], split[1:])
			result.NeedsVision = result.NeedsVision || nv
		} else if slices.Contains(result.Video.Params(), split[0]) {
			_ = result.Video.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.BackgroundRemoval.Params(), split[0]) {
			_ = result.BackgroundRemoval.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.SourceCrop.Params(), split[0]) {
			_ = result.SourceCrop.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Levels.Params(), split[0]) {
			_ = result.Levels.ParseParams(split[0], split[1:])
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
		} else if slices.Contains(result.Overlays.Params(), split[0]) {
			_ = result.Overlays.ParseParams(split[0], split[1:])
		} else if slices.Contains(result.Mask.Params(), split[0]) {
			_ = result.Mask.ParseParams(split[0], split[1:])
		} else if split[0] == "meta" {
			result.MetaOnly = true
			result.NeedsVision = true
		}

	}

	return &result, nil
}
