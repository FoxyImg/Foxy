package metadata

import (
	"slices"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
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

type DebugOptions struct {
	Faces       bool `json:"faces"`
	AllFaces    bool `json:"allFaces"`
	People      bool `json:"people"`
	AllPeople   bool `json:"allPeople"`
	OtherLabels bool `json:"otherLabels"`

	DisableSourceCache bool `json:"disableSourceCache"`
	DisableRenderCache bool `json:"disableRenderCache"`
	DisableMetaCache   bool `json:"disableMetaCache"`
}

type ImageParams struct {
	Debug DebugOptions `json:"debug"`

	MetaOnly    bool `json:"metaOnly"`
	NeedsVision bool `json:"needsVision"`

	Crop        *[]string         `json:"crop"`
	FaceIndex   *int              `json:"faceIndex"`
	PersonIndex *int              `json:"personIndex"`
	Width       *int              `json:"width"`
	Height      *int              `json:"height"`
	AspectRatio *float64          `json:"aspectRatio"`
	Gravity     *string           `json:"gravity"`
	Interesting *vips.Interesting `json:"interesting"`

	BackgroundColor ColorRGBA `json:"bgColor"`

	Rotate *int  `json:"rotate"`
	FlipH  *bool `json:"flipH"`
	FlipV  *bool `json:"flipV"`

	Brightness float64 `json:"brightness"`
	Saturation float64 `json:"saturation"`
	Hue        float64 `json:"hue"`

	ExportParams ImageExportParams `json:"export"`

	Blur *int `json:"blur"`
}

func BuildParams(pathParts []string) (*ImageParams, error) {
	result := ImageParams{
		ExportParams: ImageExportParams{
			Format:  "webp",
			Quality: 85,
		},
	}

	for _, part := range pathParts {
		split := strings.Split(part, ":")

		switch split[0] {
		// Debug
		case "debug":
			if len(split) != 2 {
				continue
			}

			types := strings.Split(split[1], ",")
			if slices.Contains(types, "faces") {
				result.NeedsVision = true
				result.Debug.Faces = true
			}

			if slices.Contains(types, "all-faces") {
				result.NeedsVision = true
				result.Debug.AllFaces = true
			}

			if slices.Contains(types, "people") {
				result.NeedsVision = true
				result.Debug.People = true
			}

			if slices.Contains(types, "all-people") {
				result.NeedsVision = true
				result.Debug.AllPeople = true
			}

			if slices.Contains(types, "other-labels") {
				result.NeedsVision = true
				result.Debug.OtherLabels = true
			}
		case "nocache":
			if len(split) != 2 {
				continue
			}

			types := strings.Split(split[1], ",")
			result.Debug.DisableSourceCache = slices.Contains(types, "source")
			result.Debug.DisableRenderCache = slices.Contains(types, "render")
			result.Debug.DisableMetaCache = slices.Contains(types, "meta")
		// Cropping/Resizing
		case "crop":
			if len(split) != 2 {
				continue
			}
			types := strings.Split(split[1], ",")
			result.Crop = &types
			result.NeedsVision = result.NeedsVision || slices.Contains(types, "face") || slices.Contains(types, "person")
		case "w":
			if len(split) != 2 {
				continue
			}

			w, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			result.Width = &w
		case "h":
			if len(split) != 2 {
				continue
			}

			h, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			result.Height = &h
		case "ar":
			if len(split) != 3 {
				continue
			}

			arw, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			arh, err := strconv.Atoi(split[2])
			if err != nil {
				continue
			}

			ar := float64(arw) / float64(arh)
			result.AspectRatio = &ar
		case "face":
			if len(split) != 2 {
				continue
			}

			f, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			result.FaceIndex = &f
		case "person":
			if len(split) != 2 {
				continue
			}

			p, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			result.PersonIndex = &p
		case "smart":
			if len(split) != 2 {
				continue
			}

			var interesting vips.Interesting
			switch split[1] {
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

			result.Interesting = &interesting
		case "bg":
			if len(split) != 2 {
				continue
			}

			c, err := ParseHexColor(split[1])
			if err != nil {
				continue
			}

			result.BackgroundColor = c
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
		// Blur
		case "blur":
			if len(split) != 2 {
				continue
			}

			b, err := strconv.Atoi(split[1])
			if err != nil {
				continue
			}

			result.Blur = &b
		default:
		}
	}

	return &result, nil
}
