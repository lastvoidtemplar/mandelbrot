package main

import (
	"image/color"
	"math"
)

func getColorSmooth(iter int, real float64, imaginary float64, maxIter int) color.Color {
	if iter == maxIter {
		return color.Black
	}

	mag := real*real + imaginary*imaginary
	mu := float64(iter) + 1 - math.Log(0.5*math.Log(mag))

	mu = float64(iter)

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

func getColorFast(iter int, maxIter int) color.RGBA {
	inten := (255 * float64(maxIter-iter) / float64(maxIter))
	return color.RGBA{
		R: uint8(inten),
		G: uint8(inten),
		B: uint8(inten),
		A: 255,
	}
}

func abs(t int) int {
	if t < 0 {
		return -t
	}
	return t
}

var Black = color.RGBA{
	R: 0,
	G: 0,
	B: 0,
	A: 255,
}

func GetColorByHue(iter int, maxIter int) color.RGBA {
	if iter == maxIter {
		return Black
	}

	const scaleFactor = 100_000

	hue := int(scaleFactor * (360.0 * float64(iter) / float64(maxIter)))

	x := scaleFactor - abs((hue/60)%(2*scaleFactor)-scaleFactor)

	xScaleBack := 255 * x / scaleFactor

	var r, g, b int
	switch {
	case hue < 60*scaleFactor:
		r, g, b = 255, xScaleBack, 0
	case hue < 120*scaleFactor:
		r, g, b = xScaleBack, 255, 0
	case hue < 180*scaleFactor:
		r, g, b = 0, 255, x*xScaleBack
	case hue < 240*scaleFactor:
		r, g, b = 0, xScaleBack, 255
	case hue < 300*scaleFactor:
		r, g, b = xScaleBack, 0, 255
	default:
		r, g, b = 255, 0, xScaleBack
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}

}
