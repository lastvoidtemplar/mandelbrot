package main

import (
	"fmt"
	"sync"
)

type AsyncQueue[T any] struct {
	startPos   int
	len        int
	arr        []T
	mutex      sync.RWMutex
	notifyChan chan struct{}
}

const StealingTreshold = 5

func NewAsyncQueue[T any](capacity int) *AsyncQueue[T] {
	return &AsyncQueue[T]{
		arr:        make([]T, capacity),
		len:        0,
		mutex:      sync.RWMutex{},
		notifyChan: make(chan struct{}, 1),
	}
}

func (q *AsyncQueue[T]) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.len
}

func (q *AsyncQueue[T]) Push(values ...T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if cap(q.arr) <= q.len+len(values) {
		newArr := make([]T, 2*cap(q.arr))
		n := copy(newArr, q.arr[q.startPos:])
		copy(newArr[n:], q.arr[:q.startPos])
		q.arr = newArr
		q.startPos = 0
	}

	for _, val := range values {
		ind := (q.startPos + q.len) % cap(q.arr)
		q.len++
		q.arr[ind] = val
	}

	if q.len > StealingTreshold {
		select {
		case q.notifyChan <- struct{}{}:
		default:
		}
	}
}

func (q *AsyncQueue[T]) Pop() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var zeroVal T
	if q.len == 0 {
		return zeroVal, false
	}

	val := q.arr[q.startPos]
	q.startPos = (q.startPos + 1) % cap(q.arr)
	q.len--
	return val, true
}

func (q *AsyncQueue[T]) Steal(batch int) ([]T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.len < StealingTreshold {
		return nil, false
	}

	count := min(q.len-StealingTreshold, batch)
	if count < 0 {
		panic(fmt.Sprintf("Count: %d, Len: %d, Batch: %d\n", count, q.len, batch))
	}
	stolenTasks := make([]T, 0, count)
	for i := 0; i < count; i++ {
		ind := (cap(q.arr) + q.startPos + q.len - 1) % cap(q.arr)
		val := q.arr[ind]
		q.len--
		stolenTasks = append(stolenTasks, val)
	}

	if q.len > StealingTreshold {
		select {
		case q.notifyChan <- struct{}{}:
		default:
		}
	}

	return stolenTasks, true
}

func (q *AsyncQueue[T]) CanSteal() <-chan struct{} {
	return q.notifyChan
}
