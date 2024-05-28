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

type BoundingBoxCropParams struct {
	Index    *int     `json:"index"`
	Padding  int      `json:"padding"`
	Zoom     *float64 `json:"zoom"`
	HGravity string   `json:"hGravity"`
	VGravity string   `json:"vGravity"`
	Largest  *bool    `json:"largest"`
	Smallest *bool    `json:"smallest"`
	Focus    bool     `json:"focus"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FocalPointOptions struct {
	X    float64  `json:"x"`
	Y    float64  `json:"y"`
	Zoom *float64 `json:"zoom"`
}

type ImageParams struct {
	Debug DebugOptions `json:"-"`

	MetaOnly    bool `json:"-"`
	NeedsVision bool `json:"vision"`

	Crop        *[]string             `json:"crop"`
	Width       *int                  `json:"width"`
	Height      *int                  `json:"height"`
	AspectRatio *float64              `json:"aspectRatio"`
	Zoom        *float64              `json:"zoom"`
	HGravity    string                `json:"hGravity"`
	VGravity    string                `json:"vGravity"`
	Face        BoundingBoxCropParams `json:"face"`
	Person      BoundingBoxCropParams `json:"person"`
	Interesting *vips.Interesting     `json:"interesting"`
	FocalPoint  FocalPointOptions     `json:"focalPoint"`

	BackgroundColor string `json:"bgColor"`

	Rotate *int  `json:"rotate"`
	FlipH  *bool `json:"flipH"`
	FlipV  *bool `json:"flipV"`

	Brightness *float64 `json:"brightness"`
	Saturation *float64 `json:"saturation"`
	Hue        *float64 `json:"hue"`

	ExportParams ImageExportParams `json:"export"`

	Blur *int `json:"blur"`
}

func NewImageParams() *ImageParams {
	return &ImageParams{
		HGravity:        "center",
		VGravity:        "center",
		BackgroundColor: "#00000000",
		Face: BoundingBoxCropParams{
			HGravity: "center",
			VGravity: "top",
			Padding:  8,
			Focus:    false,
		},
		Person: BoundingBoxCropParams{
			HGravity: "center",
			VGravity: "center",
			Padding:  0,
			Focus:    false,
		},
		FocalPoint: FocalPointOptions{
			X: 0.5,
			Y: 0.5,
		},
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
		case "zoom":
			if len(split) != 2 {
				continue
			}

			z, err := strconv.ParseFloat(split[1], 64)
			if err != nil {
				continue
			}

			result.Zoom = &z
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
		case "fp":
			if len(split) < 2 {
				continue
			}

			if split[1] == "zoom" {
				if len(split) != 3 {
					continue
				}

				z, err := strconv.ParseFloat(split[2], 64)
				if err == nil && z > 0 {
					z = z / 100.0
					result.FocalPoint.Zoom = &z
				}
			} else if len(split) == 3 {
				fpx, err := strconv.ParseFloat(split[1], 64)
				if err != nil {
					continue
				}

				fpy, err := strconv.ParseFloat(split[2], 64)
				if err != nil {
					continue
				}

				result.FocalPoint.X = fpx
				result.FocalPoint.Y = fpy
			}
		case "face":
			if len(split) < 2 {
				continue
			}

			if split[1] == "index" && len(split) == 3 {
				t := true
				if split[2] == "largest" {
					result.Face.Largest = &t
				} else if split[2] == "smallest" {
					result.Face.Smallest = &t
				} else {
					f, err := strconv.Atoi(split[2])
					if err == nil {
						result.Face.Index = &f

					}
				}
			} else if split[1] == "pad" && len(split) == 3 {
				f, err := strconv.Atoi(split[2])
				if err == nil {
					result.Face.Padding = f

				}
			} else if split[1] == "zoom" && len(split) == 3 {
				f, err := strconv.ParseFloat(split[2], 64)
				if err == nil {
					f = f / 100.0
					result.Face.Zoom = &f
				}
			} else if split[1] == "gravity" && len(split) == 4 {
				result.Face.HGravity = split[2]
				result.Face.VGravity = split[3]
			} else if split[1] == "focus" {
				result.Face.Focus = true
			}
		case "person":
			if len(split) < 2 {
				continue
			}

			if split[1] == "index" && len(split) == 3 {
				t := true
				if split[2] == "largest" {
					result.Person.Largest = &t
				} else if split[2] == "smallest" {
					result.Person.Smallest = &t
				} else {
					f, err := strconv.Atoi(split[2])
					if err == nil {
						result.Person.Index = &f

					}
				}
			} else if split[1] == "pad" && len(split) == 3 {
				f, err := strconv.Atoi(split[2])
				if err == nil {
					result.Person.Padding = f

				}
			} else if split[1] == "zoom" && len(split) == 3 {
				f, err := strconv.ParseFloat(split[2], 64)
				if err == nil {
					f = f / 100.0
					result.Person.Zoom = &f
				}
			} else if split[1] == "gravity" && len(split) == 4 {
				result.Person.HGravity = split[2]
				result.Person.VGravity = split[3]
			} else if split[1] == "focus" {
				result.Person.Focus = true
			}
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
		case "gravity":
			if len(split) == 2 {
				result.HGravity = split[1]
				result.VGravity = split[1]
			} else if len(split) == 3 {
				result.HGravity = split[1]
				result.VGravity = split[2]
			}
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
