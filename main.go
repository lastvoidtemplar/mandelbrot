package main

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	f, _ := os.Create("cpu.pprof")
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	st := time.Now()
	width := 8000
	height := 8000
	maxIter := 500
	workers := NewMandelbrotWorkerRing(12, 2000)
	tasksWg := DistributeEvenlyTasks(workers, DimensionOption{
		stX:    -2,
		stY:    -2,
		endX:   2,
		endY:   2,
		width:  width,
		height: height,
	}, GranularityOption{
		width:  400,
		height: 400,
	})

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	WorkersRun(workers, tasksWg, img, maxIter)

	log.Println("Total", time.Since(st))

	file, _ := os.Create("mandelbrot.jpg")
	jpeg.Encode(file, img, nil)
}
