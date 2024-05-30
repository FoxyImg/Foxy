package geometry

import "math"

type Size struct {
	Width  int
	Height int
}

func SizeToFillSize(innerWidth int, innerHeight int, outerWidth int, outerHeight int, preserveHeight bool) Size {
	if innerWidth <= 0 || innerHeight <= 0 {
		return Size{Width: outerWidth, Height: outerHeight}
	}

	ratioW := float64(outerWidth) / float64(innerWidth)
	ratioH := float64(outerHeight) / float64(innerHeight)

	if (ratioW > ratioH) && (preserveHeight) {
		return Size{Width: outerWidth, Height: int(math.Round(float64(innerHeight) * ratioW))}
	} else {
		return Size{Width: int(math.Round(float64(innerHeight) * ratioH)), Height: outerHeight}
	}
}

func SizeToFitSize(innerWidth int, innerHeight int, outerWidth int, outerHeight int) Size {
	if innerWidth <= 0 || innerHeight <= 0 {
		return Size{Width: outerWidth, Height: outerHeight}
	}

	if outerWidth <= 0 && outerHeight <= 0 {
		return Size{Width: 0, Height: 0}
	}

	if outerWidth <= 0 {
		if innerHeight <= outerHeight {
			return Size{Width: innerWidth, Height: innerHeight}
		}

		ratio := float64(outerHeight) / float64(innerHeight)

		return Size{Width: int(math.Round(float64(innerWidth) * ratio)), Height: outerHeight}
	}

	if outerHeight <= 0 {
		if innerWidth <= outerWidth {
			return Size{Width: innerWidth, Height: innerHeight}
		}

		ratio := float64(outerWidth) / float64(innerWidth)

		return Size{Width: outerWidth, Height: int(math.Round(float64(innerHeight) * ratio))}
	}

	ratioW := float64(outerWidth) / float64(innerWidth)
	ratioH := float64(outerHeight) / float64(innerHeight)

	if ratioW < ratioH {
		return Size{Width: outerWidth, Height: int(math.Round(float64(innerHeight) * ratioW))}
	} else {
		return Size{Width: int(math.Round(float64(innerWidth) * ratioH)), Height: outerHeight}
	}
}
