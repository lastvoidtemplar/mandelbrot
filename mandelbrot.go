package main

import (
	"image"
	"log"
	"math/bits"
	"math/rand"
	"sync"
	"time"
)

const MaxThreadsLog int = 9

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
	neighbourQueue [MaxThreadsLog](*AsyncQueue[MandelbrotTask])
}

func NewMandelbrotWorkerRing(count int) []MandelbrotWorker {
	workers := make([]MandelbrotWorker, 0, count)
	for range count {
		workers = append(workers, MandelbrotWorker{
			queue: NewAsyncQueue[MandelbrotTask](64),
		})
	}

	for i := range len(workers) {
		perv := (len(workers) + i - 1) % len(workers)
		next := (i + 1) % len(workers)
		workers[i].neighbourQueue[0] = workers[perv].queue
		workers[i].neighbourQueue[1] = workers[next].queue
		for j := 2; j < MaxThreadsLog; j++ {
			workers[i].neighbourQueue[j] = NewAsyncQueue[MandelbrotTask](64)
		}
	}

	return workers
}

func NewMandelbrotWorkerHyperCube(count int) []MandelbrotWorker {
	workers := make([]MandelbrotWorker, 0, count)
	for range count {
		workers = append(workers, MandelbrotWorker{
			queue: NewAsyncQueue[MandelbrotTask](64),
		})
	}

	k := bits.Len(uint(count - 1))

	for i := range count {
		for b := 0; b < MaxThreadsLog; b++ {
			if b < k {
				neighbor := i ^ (1 << b)
				if neighbor >= count {
					neighbor = rand.Intn(count)
				}
				workers[i].neighbourQueue[b] = workers[neighbor].queue
			} else {
				workers[i].neighbourQueue[b] = NewAsyncQueue[MandelbrotTask](64)
			}
		}
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

func WorkersRun(workers []MandelbrotWorker, tasksWg *sync.WaitGroup, img *image.RGBA, maxIter int, stealing bool) {
	p := len(workers)
	for i := 0; i < p; i++ {
		workers[i].img = img
	}

	doneCh := make(chan struct{}, p)

	for i := 0; i < p; i++ {
		go workers[i].Run(maxIter, tasksWg, doneCh, stealing)
	}

	tasksWg.Wait()

	for i := 0; i < p; i++ {
		doneCh <- struct{}{}
	}
}

const stealBatchSize = 10

func (worker *MandelbrotWorker) Run(maxIter int, taskWg *sync.WaitGroup, doneCh <-chan struct{}, stealing bool) {
	t, _ := time.ParseDuration("0s")

	stX := worker.dimension.stX
	stY := worker.dimension.stY
	stepX := (worker.dimension.endX - worker.dimension.stX) / float64(worker.dimension.width)
	stepY := (worker.dimension.endY - worker.dimension.stY) / float64(worker.dimension.height)
	for {
		task, ok := worker.queue.Pop()
		if ok {
			st := time.Now()
			worker.process(task, stX, stY, stepX, stepY, maxIter)
			taskWg.Done()
			d := time.Since(st)
			t += d
			continue
		}

		if stealing {
			select {
			case <-worker.neighbourQueue[0].CanSteal():
				worker.steal(0)
			case <-worker.neighbourQueue[1].CanSteal():
				worker.steal(1)
			case <-worker.neighbourQueue[2].CanSteal():
				worker.steal(2)
			case <-worker.neighbourQueue[3].CanSteal():
				worker.steal(3)
			case <-worker.neighbourQueue[4].CanSteal():
				worker.steal(4)
			case <-worker.neighbourQueue[5].CanSteal():
				worker.steal(5)
			case <-worker.neighbourQueue[6].CanSteal():
				worker.steal(6)
			case <-worker.neighbourQueue[7].CanSteal():
				worker.steal(7)
			case <-worker.neighbourQueue[8].CanSteal():
				worker.steal(8)
			case <-doneCh:
				log.Println(t)
			}
			continue
		}

		<-doneCh
		log.Println(t)

	}
}

func (worker *MandelbrotWorker) steal(neighbor int) {
	l := worker.neighbourQueue[neighbor].Len()
	tasks, ok := worker.neighbourQueue[neighbor].Steal(max(stealBatchSize, l/3))
	if ok {
		worker.queue.Push(tasks...)
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
