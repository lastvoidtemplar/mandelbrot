package main

import (
	"flag"
	"image"
	"image/jpeg"
	"log"
	"os"
	"sync"
	"time"
)

type Topology = string

const (
	Ring      Topology = "ring"
	HyperCube Topology = "hypercube"
)

type Distribute = string

const (
	Evenly   Distribute = "evenly"
	Dump     Distribute = "dump"
	Randomly Distribute = "randomly"
)

func main() {
	var maxIter int
	var threads int
	var topology Topology
	var stX float64
	var stY float64
	var endX float64
	var endY float64
	var resultWidth int
	var resultHeight int
	var granWidth int
	var granHeight int
	var distribute Distribute
	var createImage bool
	var stealing bool

	flag.IntVar(&maxIter, "iter", 500, "Max iterations for each number")
	flag.IntVar(&threads, "threads", 2, "The number of threads that will be spawn")
	flag.StringVar(&topology, "topology", "hypercube", "The topology of the workers(ring, hypercube)")
	flag.Float64Var(&stX, "st-x", -2.0, "The cordinate of x-axis of the starting point")
	flag.Float64Var(&stY, "st-y", -2.0, "The cordinate of y-axis of the starting point")
	flag.Float64Var(&endX, "end-x", 2.0, "The cordinate of x-axis of the ending point")
	flag.Float64Var(&endY, "end-y", 2.0, "The cordinate of y-axis of the ending point")
	flag.IntVar(&resultWidth, "result-width", 8000, "The width of the result image")
	flag.IntVar(&resultHeight, "result-height", 8000, "The height of the result image")
	flag.IntVar(&granWidth, "gran-width", 400, "The width of the granularity of a task")
	flag.IntVar(&granHeight, "gran-height", 400, "The height of the granularity of a task")
	flag.StringVar(&distribute, "distribute", "dump", "The way task are distribute(evenly, dump, randomly)")
	flag.BoolVar(&createImage, "export", false, `Option to export the image ""mandelbrot.jpg`)
	flag.BoolVar(&stealing, "no-stealing", false, `Option to disable stealing`)

	flag.Parse()

	stealing = !stealing

	st := time.Now()

	var workers []MandelbrotWorker
	switch topology {
	case Ring:
		workers = NewMandelbrotWorkerRing(threads)
	case HyperCube:
		workers = NewMandelbrotWorkerHyperCube(threads)
	default:
		log.Fatalf("Invalid topology type")
	}

	var tasksWg *sync.WaitGroup
	switch distribute {
	case Evenly:
		tasksWg = DistributeEvenlyTasks(workers, DimensionOption{
			stX:    stX,
			stY:    stY,
			endX:   endX,
			endY:   endY,
			width:  resultWidth,
			height: resultHeight,
		}, GranularityOption{
			width:  granWidth,
			height: granHeight,
		})
	case Dump:
		tasksWg = DistributeTasksByDumping(workers, DimensionOption{
			stX:    stX,
			stY:    stY,
			endX:   endX,
			endY:   endY,
			width:  resultWidth,
			height: resultHeight,
		}, GranularityOption{
			width:  granWidth,
			height: granHeight,
		})
	case Randomly:
		tasksWg = DistributeRandomlyTasks(workers, DimensionOption{
			stX:    stX,
			stY:    stY,
			endX:   endX,
			endY:   endY,
			width:  resultWidth,
			height: resultHeight,
		}, GranularityOption{
			width:  granWidth,
			height: granHeight,
		})
	default:
		log.Fatalf("Invalid distribute type")
	}

	img := image.NewRGBA(image.Rect(0, 0, resultWidth, resultHeight))
	WorkersRun(workers, tasksWg, img, maxIter, stealing)

	log.Println("Total", time.Since(st))

	if createImage {
		file, _ := os.Create("mandelbrot.jpg")
		jpeg.Encode(file, img, nil)
	}
}
