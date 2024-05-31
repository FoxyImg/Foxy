package metadata

import (
	"errors"
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

type Rect struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type FocalPointOptions struct {
	X    float64  `json:"x"`
	Y    float64  `json:"y"`
	Zoom *float64 `json:"zoom"`
}

type BorderOptions struct {
	Color  string `json:"color"`
	Left   int    `json:"left"`
	Top    int    `json:"top"`
	Right  int    `json:"right"`
	Bottom int    `json:"bottom"`
}

type RedactOptions struct {
	Faces        []int   `json:"faces"`
	People       []int   `json:"people"`
	Regions      []Rect  `json:"regions"`
	Blur         int     `json:"blur"`
	BlurMask     int     `json:"blurMask"`
	ExpandMask   int     `json:"expandMask"`
	UseColor     bool    `json:"useColor"`
	Color        *string `json:"color"`
	Pixelate     int     `json:"pixelate"`
	PixelateMask int     `json:"pixelateMask"`
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

	Padding *BorderOptions `json:"padding"`
	Border  *BorderOptions `json:"border"`

	Redact *RedactOptions `json:"redact"`

	Rotate *int  `json:"rotate"`
	FlipH  *bool `json:"flipH"`
	FlipV  *bool `json:"flipV"`

	Brightness *float64 `json:"brightness"`
	Saturation *float64 `json:"saturation"`
	Hue        *float64 `json:"hue"`

	Blur *int `json:"blur"`

	ExportParams ImageExportParams `json:"export"`
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

func parseBorderOptions(split []string) (*BorderOptions, error) {
	if len(split) <= 2 || len(split) == 5 {
		return nil, errors.New("invalid border options")
	}

	var options = BorderOptions{
		Color:  split[1],
		Left:   0,
		Top:    0,
		Right:  0,
		Bottom: 0,
	}

	if len(split) == 3 {
		p, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}

		options.Left = p
		options.Top = p
		options.Right = p
		options.Bottom = p
	} else if len(split) == 4 {
		h, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}

		v, err := strconv.Atoi(split[3])
		if err != nil {
			return nil, err
		}

		options.Left = h
		options.Top = v
		options.Right = h
		options.Bottom = v
	} else if len(split) == 6 {
		l, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}

		t, err := strconv.Atoi(split[3])
		if err != nil {
			return nil, err
		}

		r, err := strconv.Atoi(split[4])
		if err != nil {
			return nil, err
		}

		b, err := strconv.Atoi(split[5])
		if err != nil {
			return nil, err
		}

		options.Left = l
		options.Top = t
		options.Right = r
		options.Bottom = b
	}

	return &options, nil
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
		case "pad": // pad:<color>:<left>:<top>:<right>:<bottom>
			options, err := parseBorderOptions(split)
			if err != nil {
				continue
			}

			result.Padding = options
		case "border":
			options, err := parseBorderOptions(split)
			if err != nil {
				continue
			}

			result.Border = options
		case "redact":
			if len(split) < 3 {
				continue
			}

			result.NeedsVision = true

			if result.Redact == nil {
				result.Redact = &RedactOptions{
					Faces:        []int{},
					People:       []int{},
					Regions:      []Rect{},
					Blur:         0,
					UseColor:     false,
					Pixelate:     0,
					BlurMask:     0,
					ExpandMask:   0,
					PixelateMask: 0,
				}
			}

			if split[1] == "faces" {
				faces := strings.Split(split[2], ",")
				if slices.Contains(faces, "all") {
					result.Redact.Faces = append(result.Redact.Faces, -1)
				} else {
					for _, face := range faces {
						f, err := strconv.Atoi(face)
						if err != nil {
							continue
						}

						result.Redact.Faces = append(result.Redact.Faces, f)
					}
				}
			} else if split[1] == "people" {
				people := strings.Split(split[2], ",")
				if slices.Contains(people, "all") {
					result.Redact.People = append(result.Redact.People, -1)
				} else {
					for _, person := range people {
						p, err := strconv.Atoi(person)
						if err != nil {
							continue
						}

						result.Redact.People = append(result.Redact.People, p)
					}
				}
			} else if split[1] == "region" {
				regionParts := strings.Split(split[2], ",")
				if len(regionParts) != 4 {
					continue
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

				result.Redact.Regions = append(result.Redact.Regions, r)
			} else if split[1] == "blur" {
				b, _ := strconv.Atoi(split[2])
				result.Redact.Blur = b
			} else if split[1] == "color" {
				if len(split) < 3 {
					continue
				}

				result.Redact.UseColor = true
				result.Redact.Color = &split[2]
			} else if split[1] == "pixelate" {
				p, _ := strconv.Atoi(split[2])
				result.Redact.Pixelate = p
			} else if split[1] == "mask" {
				if len(split) < 4 {
					continue
				}

				if split[2] == "blur" {
					b, err := strconv.Atoi(split[3])
					if err == nil {
						result.Redact.BlurMask = b
					}
				} else if split[2] == "expand" {
					b, err := strconv.Atoi(split[3])
					if err == nil {
						result.Redact.ExpandMask = b
					}
				} else if split[2] == "pixelate" {
					b, err := strconv.Atoi(split[3])
					if err == nil {
						result.Redact.PixelateMask = b
					}
				}
			}
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
