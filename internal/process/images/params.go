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

type ImageParams struct {
	DebugFaces       bool `json:"debugFaces"`
	DebugAllFaces    bool `json:"debugAllFaces"`
	DebugPeople      bool `json:"debugPeople"`
	DebugAllPeople   bool `json:"debugAllPeople"`
	DebugOtherLabels bool `json:"debugOtherLabels"`
	MetaOnly         bool `json:"metaOnly"`
	NeedsVision      bool `json:"needsVision"`

	Crop        *[]string `json:"crop"`
	FaceIndex   *int      `json:"faceIndex"`
	PersonIndex *int      `json:"personIndex"`
	Width       *int      `json:"width"`
	Height      *int      `json:"height"`
	Gravity     *string   `json:"gravity"`

	Rotate *int  `json:"rotate"`
	FlipH  *bool `json:"flipH"`
	FlipV  *bool `json:"flipV"`

	Brightness float64 `json:"brightness"`
	Saturation float64 `json:"saturation"`
	Hue        float64 `json:"hue"`

	ExportParams ImageExportParams `json:"export"`

	Blur *int `json:"blur"`

	DisableSourceCache bool `json:"disableSourceCache"`
	DisableRenderCache bool `json:"disableRenderCache"`
	DisableMetaCache   bool `json:"disableMetaCache"`
}

func BuildParams(pathParts []string) (*ImageParams, error) {
	result := ImageParams{
		ExportParams: ImageExportParams{
			Format:  "jpg",
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
				result.DebugFaces = true
			}

			if slices.Contains(types, "all-faces") {
				result.NeedsVision = true
				result.DebugAllFaces = true
			}

			if slices.Contains(types, "people") {
				result.NeedsVision = true
				result.DebugPeople = true
			}

			if slices.Contains(types, "all-people") {
				result.NeedsVision = true
				result.DebugAllPeople = true
			}

			if slices.Contains(types, "other-labels") {
				result.NeedsVision = true
				result.DebugOtherLabels = true
			}
		case "nocache":
			if len(split) != 2 {
				continue
			}

			types := strings.Split(split[1], ",")
			result.DisableSourceCache = slices.Contains(types, "source")
			result.DisableRenderCache = slices.Contains(types, "render")
			result.DisableMetaCache = slices.Contains(types, "meta")
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
