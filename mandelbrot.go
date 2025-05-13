package main

import (
	"image"
	"log"
	"math/rand"
	"sync"
	"time"
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

func DistributeEvenlyTasks(workers []MandelbrotWorker, dimension DimensionOption, granularity GranularityOption) *sync.WaitGroup {
	ind := 0
	p := len(workers)

	for i := 0; i < p; i++ {
		workers[i].dimension = dimension
	}

	tasksWg := &sync.WaitGroup{}
	for xInd := 0; xInd < dimension.width; xInd += granularity.width {
		for yInd := 0; yInd < dimension.height; yInd += granularity.height {
			workers[ind].queue.Push(MandelbrotTask{
				stXInd:  xInd,
				stYInd:  yInd,
				endXInd: min(xInd+granularity.width, dimension.width),
				endYInd: min(yInd+granularity.height, dimension.height),
			})
			tasksWg.Add(1)
			ind = (ind + 1) % p
		}
	}
	return tasksWg
}

func DistributeTasksByDumping(workers []MandelbrotWorker, dimension DimensionOption, granularity GranularityOption) *sync.WaitGroup {
	p := len(workers)

	for i := 0; i < p; i++ {
		workers[i].dimension = dimension
	}

	tasksWg := &sync.WaitGroup{}
	for xInd := 0; xInd < dimension.width; xInd += granularity.width {
		for yInd := 0; yInd < dimension.height; yInd += granularity.height {
			workers[0].queue.Push(MandelbrotTask{
				stXInd:  xInd,
				stYInd:  yInd,
				endXInd: min(xInd+granularity.width, dimension.width),
				endYInd: min(yInd+granularity.height, dimension.height),
			})
			tasksWg.Add(1)
		}
	}
	return tasksWg
}

func DistributeRandomlyTasks(workers []MandelbrotWorker, dimension DimensionOption, granularity GranularityOption) *sync.WaitGroup {
	p := len(workers)

	for i := 0; i < p; i++ {
		workers[i].dimension = dimension
	}

	tasksWg := &sync.WaitGroup{}
	for xInd := 0; xInd < dimension.width; xInd += granularity.width {
		for yInd := 0; yInd < dimension.height; yInd += granularity.height {
			workers[rand.Intn(p)].queue.Push(MandelbrotTask{
				stXInd:  xInd,
				stYInd:  yInd,
				endXInd: min(xInd+granularity.width, dimension.width),
				endYInd: min(yInd+granularity.height, dimension.height),
			})
			tasksWg.Add(1)
		}
	}
	return tasksWg
}

func WorkersRun(workers []MandelbrotWorker, tasksWg *sync.WaitGroup, img *image.RGBA, maxIter int) {
	p := len(workers)
	for i := 0; i < p; i++ {
		workers[i].img = img
	}

	doneCh := make(chan struct{}, p)

	for i := 0; i < p; i++ {
		go workers[i].Run(maxIter, tasksWg, doneCh)
	}

	tasksWg.Wait()

	for i := 0; i < p; i++ {
		doneCh <- struct{}{}
	}
}

const stealBatchSize = 10

func (worker *MandelbrotWorker) Run(maxIter int, taskWg *sync.WaitGroup, doneCh <-chan struct{}) {
	st := time.Now()

	stX := worker.dimension.stX
	stY := worker.dimension.stY
	stepX := (worker.dimension.endX - worker.dimension.stX) / float64(worker.dimension.width)
	stepY := (worker.dimension.endY - worker.dimension.stY) / float64(worker.dimension.height)

	for {
		task, ok := worker.queue.Pop()
		if ok {
			worker.process(task, stX, stY, stepX, stepY, maxIter)
			taskWg.Done()
			continue
		}

		select {
		case <-worker.neighbourQueue[0].CanSteal():
			l := worker.neighbourQueue[0].Len()
			tasks, ok := worker.neighbourQueue[0].Steal(max(stealBatchSize, l/3))
			if ok {
				worker.queue.Push(tasks...)
			}
		case <-worker.neighbourQueue[1].CanSteal():
			l := worker.neighbourQueue[1].Len()
			tasks, ok := worker.neighbourQueue[1].Steal(max(stealBatchSize, l/3))
			if ok {
				worker.queue.Push(tasks...)
			}

		case <-doneCh:
			log.Println(time.Since(st))
		}

	}
}

func (worker *MandelbrotWorker) process(task MandelbrotTask, stX float64, stY float64, stepX float64, stepY float64, maxIter int) {
	for xInd := task.stXInd; xInd < task.endXInd; xInd++ {
		x := stX + float64(xInd)*stepX
		for yInd := task.stYInd; yInd < task.endYInd; yInd++ {
			y := stY + float64(yInd)*stepY

			real := 0.0
			imaginary := 0.0
			i := 0
			for ; i < maxIter; i++ {
				real, imaginary = real*real-imaginary*imaginary+x, 2*real*imaginary+y
				if real*real+imaginary*imaginary > 4 {
					break
				}
			}
			worker.img.SetRGBA(xInd, yInd, GetColorByHue(i, maxIter))
		}
	}
}
