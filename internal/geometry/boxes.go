package geometry

import (
	"math"
)

type Box struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func CalcBoxBounds(boxes []Box) Box {
	var boxLeft = math.MaxFloat64
	var boxTop = math.MaxFloat64
	var boxRight float64 = 0
	var boxBottom float64 = 0
	for _, box := range boxes {
		boxLeft = math.Min(boxLeft, box.Left)
		boxTop = math.Min(boxTop, box.Top)
		boxRight = math.Max(boxRight, box.Left+box.Width)
		boxBottom = math.Max(boxBottom, box.Top+box.Height)
	}

	return Box{
		Left:   boxLeft,
		Top:    boxTop,
		Width:  boxRight - boxLeft,
		Height: boxBottom - boxTop,
	}
}
