package main

import (
	"image/color"
	"math"
	"math/cmplx"
)

func getColor(iter int, z complex128, maxIter int) color.Color {
	if iter == maxIter {
		return color.Black
	}

	mag := cmplx.Abs(z)
	mu := float64(iter) + 1 - math.Log(math.Log2(mag))

	hue := 360.0 * mu / float64(maxIter)
	sat := 1.0
	val := 1.0

	return hsvToRGB(hue, sat, val)
}

func hsvToRGB(h, s, v float64) color.RGBA {
	if s == 0.0 {
		r := uint8(v * 255)
		return color.RGBA{r, r, r, 255}
	}

	h = math.Mod(h, 360)
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		uint8((r + m) * 255),
		uint8((g + m) * 255),
		uint8((b + m) * 255),
		255,
	}
}
