package main

import (
	"image"
	"math/cmplx"
	"sync"
)

type MandelbrotTask struct {
	stXInd  int
	stYInd  int
	endXInd int
	endYInd int
}

type DimensionOption struct {
	stX    float64
	stY    float64
	endX   float64
	endY   float64
	width  int
	height int
}

type MandelbrotWorker struct {
	img            *image.RGBA
	dimension      DimensionOption
	queue          *AsyncQueue[MandelbrotTask]
	neighbourQueue [2](*AsyncQueue[MandelbrotTask])
}

func NewMandelbrotWorkerRing(count int, queueCapacity int) []MandelbrotWorker {
	workers := make([]MandelbrotWorker, 0, count)
	for range count {
		workers = append(workers, MandelbrotWorker{
			queue: NewAsyncQueue[MandelbrotTask](queueCapacity),
		})
	}

	for i := range len(workers) {
		perv := (len(workers) + i - 1) % len(workers)
		next := (i + 1) % len(workers)
		workers[i].neighbourQueue[0] = workers[perv].queue
		workers[i].neighbourQueue[1] = workers[next].queue
	}

	return workers
}

type GranularityOption struct {
	width  int
	height int
}

func DistributeTasks(workers []MandelbrotWorker, dimension DimensionOption, granularity GranularityOption) {
	ind := 0
	p := len(workers)

	for i := 0; i < p; i++ {
		workers[i].dimension = dimension
	}

	for xInd := 0; xInd < dimension.width; xInd += granularity.width {
		for yInd := 0; yInd < dimension.height; yInd += granularity.height {
			workers[ind].queue.Push(MandelbrotTask{
				stXInd:  xInd,
				stYInd:  yInd,
				endXInd: min(xInd+granularity.width, dimension.width),
				endYInd: min(yInd+granularity.height, dimension.height),
			})
			ind = (ind + 1) % p
		}
	}
}

func WorkersRun(workers []MandelbrotWorker, img *image.RGBA, maxIter int) {
	wg := &sync.WaitGroup{}

	p := len(workers)
	for i := 0; i < p; i++ {
		workers[i].img = img
	}

	wg.Add(p)
	for i := 0; i < p; i++ {
		go workers[i].Run(maxIter, wg)
	}

	wg.Wait()
}

func (worker *MandelbrotWorker) Run(maxIter int, wg *sync.WaitGroup) {
	defer wg.Done()

	stX := worker.dimension.stX
	stY := worker.dimension.stY
	stepX := (worker.dimension.endX - worker.dimension.stX) / float64(worker.dimension.width)
	stepY := (worker.dimension.endY - worker.dimension.stY) / float64(worker.dimension.height)

	for {
		task, ok := worker.queue.Pop()
		if !ok {
			return
		}

		for xInd := task.stXInd; xInd < task.endXInd; xInd++ {
			x := stX + float64(xInd)*stepX
			for yInd := task.stYInd; yInd < task.endYInd; yInd++ {
				y := stY + float64(yInd)*stepY

				c := complex(x, y)
				z := complex(0, 0)
				i := 0
				for ; i < maxIter; i++ {
					z = z*z + c
					mag := cmplx.Abs(z)
					if mag > 2 {
						break
					}
				}
				worker.img.Set(xInd, yInd, getColor(i, z, maxIter))
			}
		}
	}
}
