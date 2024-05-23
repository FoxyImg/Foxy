package metadata

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

type Metadata struct {
	Faces            []Face  `json:"faces"`
	Labels           []Label `json:"labels"`
	People           []Label `json:"people"`
	ModerationLabels []Label `json:"moderationLabels"`
}

func HasPersonLabel(labels []Label) bool {
	for _, label := range labels {
		if slices.Contains(PersonLabels, label.Name) {
			return true
		}
	}

	return false
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

func CalcPersonBounds(labels []Label) geometry.Box {
	var boxes []geometry.Box
	for _, label := range labels {
		if !slices.Contains(PersonLabels, label.Name) {
			continue
		}

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
