package main

import "sync"

type AsyncQueue[T any] struct {
	startPos int
	len      int
	arr      []T
	mutex    sync.RWMutex
}

func NewAsyncQueue[T any](capacity int) *AsyncQueue[T] {
	return &AsyncQueue[T]{
		arr:   make([]T, capacity),
		mutex: sync.RWMutex{},
	}
}

func (q *AsyncQueue[T]) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.len
}

func (q *AsyncQueue[T]) Push(val T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if cap(q.arr) == q.len {
		newArr := make([]T, 2*cap(q.arr))
		n := copy(newArr, q.arr[q.startPos:])
		copy(newArr[n:], q.arr[:q.startPos])
		q.arr = newArr
		q.startPos = 0
	}

	ind := (q.startPos + q.len) % cap(q.arr)
	q.len++
	q.arr[ind] = val
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

func (q *AsyncQueue[T]) Steal() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var zeroVal T
	if q.len == 0 {
		return zeroVal, false
	}

	ind := (cap(q.arr) + q.startPos + q.len - 1) % cap(q.arr)
	val := q.arr[ind]
	q.len--
	return val, true
}
