package main

import (
	"fmt"
	"image"
	"testing"
)

func TestMandelbrot(t *testing.T) {
	workers := NewMandelbrotWorkerRing(4, 4)
	DistributeTasks(workers, DimensionOption{
		stX:    -2,
		stY:    -2,
		endX:   2,
		endY:   2,
		width:  4000,
		height: 4000,
	}, GranularityOption{
		width:  100,
		height: 1,
	})
	_ = workers
}

func runMandelbrot(p int) {
	width := 8000
	height := 8000
	maxIter := 2000
	workers := NewMandelbrotWorkerRing(p, 4)
	DistributeTasks(workers, DimensionOption{
		stX:    -2,
		stY:    -2,
		endX:   2,
		endY:   2,
		width:  width,
		height: height,
	}, GranularityOption{
		width:  100,
		height: 1,
	})

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	WorkersRun(workers, img, maxIter)
}

func BenchmarkMandelbrot(b *testing.B) {
	ps := []int{1, 2, 3, 4, 6, 8, 10, 12}

	for _, p := range ps {
		b.Run(fmt.Sprintf("p=%d", p), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				runMandelbrot(p)
			}
		})
	}
}
