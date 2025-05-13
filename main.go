package main

import (
	"image"
	"image/jpeg"
	"os"
	"runtime/pprof"
)

func main() {
	f, _ := os.Create("cpu.pprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	width := 8000
	height := 8000
	maxIter := 2000
	workers := NewMandelbrotWorkerRing(16, 2000)
	DistributeTasks(workers, DimensionOption{
		stX:    -2,
		stY:    -2,
		endX:   2,
		endY:   2,
		width:  width,
		height: height,
	}, GranularityOption{
		width:  8000,
		height: 1,
	})

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	WorkersRun(workers, img, maxIter)

	file, _ := os.Create("mandelbrot.jpg")
	jpeg.Encode(file, img, nil)
}
