package vision

import (
	"foxy/internal/geometry"
	"math"
	"slices"
)

var PersonLabels = []string{
	"Adult",
	"Angel",
	"Body Part",
	"Boy",
	"Child",
	"Cupid",
	"Elf",
	"Face",
	"Female",
	"Girl",
	"Hippie",
	"Lady",
	"Male",
	"Man",
	"Military Officer",
	"Monk",
	"People",
	"Person",
	"Senior Citizen",
	"Teen",
	"Torso",
	"Tourist",
	"Troop",
	"Woman",
}

type AgeRange struct {
	Low  int64 `json:"low"`
	High int64 `json:"high"`
}

type Face struct {
	Box        geometry.Box `json:"box"`
	Gender     *string      `json:"gender"`
	Rotation   float64      `json:"rotation"`
	Age        *AgeRange    `json:"age"`
	Confidence float64      `json:"confidence"`
}

type Label struct {
	Name       string        `json:"name"`
	Box        *geometry.Box `json:"box"`
	Confidence float64       `json:"confidence"`
}

type UsedColor struct {
	Used float64 `json:"used"`
	R    int     `json:"r"`
	G    int     `json:"g"`
	B    int     `json:"b"`
	L    float64 `json:"l"`
}

type DominantColors struct {
	Lightest *UsedColor  `json:"lightest"`
	Darkest  *UsedColor  `json:"darkest"`
	Colors   []UsedColor `json:"colors"`
}

type Metadata struct {
	Width            int            `json:"width"`
	Height           int            `json:"height"`
	Faces            []Face         `json:"faces"`
	Labels           []Label        `json:"labels"`
	People           []Label        `json:"people"`
	ModerationLabels []Label        `json:"moderationLabels"`
	DominantColors   DominantColors `json:"dominantColors"`
}

func cropLabels(labels []Label, sw, sh, x, y, width, height float64) []Label {
	px := sw * x
	py := sh * y
	pw := sw * width
	ph := sh * height

	newLabels := []Label{}
	for _, label := range labels {
		if label.Box == nil {
			newLabels = append(newLabels, label)
			continue
		}

		fx := label.Box.Left * sw
		fy := label.Box.Top * sh
		fw := label.Box.Width * sw
		fh := label.Box.Height * sh

		label.Box.Left = (fx - px) / pw
		label.Box.Top = (fy - py) / ph
		label.Box.Width = fw / pw
		label.Box.Height = fh / ph

		if label.Box.Left+label.Box.Width < 0 {
			continue
		} else if label.Box.Left > 1.0 {
			continue
		} else if label.Box.Top+label.Box.Height < 0 {
			continue
		} else if label.Box.Top > 1.0 {
			continue
		}

		newLabels = append(newLabels, label)
	}

	return newLabels
}

func (params *Metadata) SourceCrop(sw, sh, x, y, width, height float64) error {
	px := sw * x
	py := sh * y
	pw := sw * width
	ph := sh * height

	newFaces := []Face{}
	for _, face := range params.Faces {
		fx := face.Box.Left * sw
		fy := face.Box.Top * sh
		fw := face.Box.Width * sw
		fh := face.Box.Height * sh

		face.Box.Left = (fx - px) / pw
		face.Box.Top = (fy - py) / ph
		face.Box.Width = fw / pw
		face.Box.Height = fh / ph

		if face.Box.Left+face.Box.Width < 0 {
			continue
		} else if face.Box.Left > 1.0 {
			continue
		} else if face.Box.Top+face.Box.Height < 0 {
			continue
		} else if face.Box.Top > 1.0 {
			continue
		}

		newFaces = append(newFaces, face)
	}
	params.Faces = newFaces

	params.Labels = cropLabels(params.Labels, sw, sh, x, y, width, height)
	params.People = cropLabels(params.People, sw, sh, x, y, width, height)
	params.ModerationLabels = cropLabels(params.ModerationLabels, sw, sh, x, y, width, height)

	return nil
}

func CalcFacesBounds(faces []Face) geometry.Box {
	var boxes []geometry.Box
	for _, face := range faces {
		boxes = append(boxes, face.Box)
	}

	return geometry.CalcBoxBounds(boxes)
}

func CalcLabelsBounds(labels []Label) geometry.Box {
	var boxes []geometry.Box
	for _, label := range labels {
		if label.Box != nil {
			boxes = append(boxes, *label.Box)
		}
	}

	if len(boxes) == 0 {
		return geometry.Box{
			Left:   0,
			Top:    0,
			Width:  0,
			Height: 0,
		}
	}

	return geometry.CalcBoxBounds(boxes)
}

func GetPersonLabels(labels []Label) []Label {
	var res []Label

	for _, label := range labels {
		if label.Box == nil {
			continue
		}

		if slices.Contains(PersonLabels, label.Name) {
			res = append(res, label)
		}
	}

	return CombineBoxLabels(res)
}

func CombineBoxLabels(labels []Label) []Label {
	var res []Label

	for _, label := range labels {
		if label.Box == nil {
			continue
		}

		foundMatch := false
		for idx, resLabel := range res {
			lbl := int(math.Floor(label.Box.Left * 100))
			lbt := int(math.Floor(label.Box.Top * 100))
			lbw := int(math.Floor(label.Box.Width * 100))
			lbh := int(math.Floor(label.Box.Height * 100))

			rbl := int(math.Floor(resLabel.Box.Left * 100))
			rbt := int(math.Floor(resLabel.Box.Top * 100))
			rbw := int(math.Floor(resLabel.Box.Width * 100))
			rbh := int(math.Floor(resLabel.Box.Height * 100))

			if lbl == rbl && lbt == rbt && lbw == rbw && lbh == rbh {
				foundMatch = true
				res[idx].Name = res[idx].Name + "/" + label.Name
			}
		}

		if !foundMatch {
			res = append(res, label)
		}
	}

	return res
}

func FilterNonPersonLabels(labels []Label) []Label {
	var res []Label

	for _, label := range labels {
		if label.Box == nil {
			continue
		}

		if !slices.Contains(PersonLabels, label.Name) {
			res = append(res, label)
		}
	}

	return res
}
