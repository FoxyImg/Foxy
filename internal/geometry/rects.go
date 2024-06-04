package geometry

import "math"

type PointFloat64 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type SizeFloat64 struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type RectFloat64 struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (size SizeFloat64) Size() Size {
	return Size{
		Width:  int(math.Round(size.Width)),
		Height: int(math.Round(size.Height)),
	}
}

func (point PointFloat64) RotatePoint(clockwise bool, anchor PointFloat64, degrees float64) {
	radians := degrees * (math.Pi / 180.0)
	s := math.Sin(radians)
	c := math.Cos(radians)

	// translate point back to origin:
	point.X -= anchor.X
	point.Y -= anchor.Y

	// rotate point
	if clockwise {
		point.X = point.X*c - point.Y*s
		point.Y = point.X*s + point.Y*c
	} else {
		point.X = point.X*c + point.Y*s
		point.Y = -point.X*s + point.Y*c
	}

	// translate point back:
	point.X += anchor.X
	point.Y += anchor.Y
}

func (r RectFloat64) Points() [4]PointFloat64 {
	return [4]PointFloat64{
		{
			X: r.X,
			Y: r.Y,
		},
		{
			X: r.X + r.Width,
			Y: r.Y,
		},
		{
			X: r.X + r.Width,
			Y: r.Y + r.Height,
		},
		{
			X: r.X,
			Y: r.Y + r.Height,
		},
	}
}

func (r RectFloat64) RotatePoints(degrees float64) [4]PointFloat64 {
	if r.Width == 0 || r.Height == 0 {
		return r.Points()
	}

	anchor := PointFloat64{
		X: r.X + r.Width/2.0,
		Y: r.Y + r.Height/2.0,
	}

	points := r.Points()
	for _, point := range points {
		point.RotatePoint(true, anchor, degrees)
	}

	return points
}

func (r RectFloat64) SizeThatFitsRotatedRect(degrees float64) Size {
	for degrees < 0 {
		degrees += 360
	}

	radians := degrees * math.Pi / 180.0

	var degt float64
	if degrees <= 90 {
		degt = radians
	} else if degrees <= 180 {
		degt = (180 - degrees) * math.Pi / 180.0
	} else if degrees <= 270 {
		degt = (degrees - 180) * math.Pi / 180.0
	} else {
		degt = (360 - degrees) * math.Pi / 180.0
	}

	sint := math.Sin(degt)
	cost := math.Cos(degt)

	h1 := r.Height * r.Height / (r.Width*sint + r.Height*cost)
	h2 := r.Height * r.Width / (r.Width*cost + r.Height*sint)
	hh := math.Min(h1, h2)
	ww := hh * r.Width / r.Height

	return Size{
		Width:  int(math.Round(ww)),
		Height: int(math.Round(hh)),
	}
}
