package main

import "testing"

func TestAsyncQueue1(t *testing.T) {
	q := NewAsyncQueue[int](4)

	if q.Len() != 0 {
		t.Fatalf("Expected %d, but got %d\n", 0, q.Len())
	}

	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	q.Push(5)

	for i := range 3 {
		val, ok := q.Pop()
		if !ok {
			t.Errorf("Expected %t, but got %t\n", true, ok)
		}

		if val != i+1 {
			t.Errorf("Expected %d, but got %d\n", i+1, val)
		}
	}

	q.Push(6)
	q.Push(7)
	q.Push(8)
	q.Push(9)
	q.Push(10)

	for i := range 7 {
		val, ok := q.Pop()
		if !ok {
			t.Errorf("Expected %t, but got %t\n", true, ok)
		}

		if val != i+4 {
			t.Errorf("Expected %d, but got %d\n", i+4, val)
		}
	}
}

func TestAsyncQueue2(t *testing.T) {
	q := NewAsyncQueue[int](4)

	if q.Len() != 0 {
		t.Fatalf("Expected %d, but got %d\n", 0, q.Len())
	}

	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	q.Push(5)

	for i := range 3 {
		val, ok := q.Steal()
		if !ok {
			t.Errorf("Expected %t, but got %t\n", true, ok)
		}

		if val != 5-i {
			t.Errorf("Expected %d, but got %d\n", 5-i, val)
		}
	}

	q.Push(6)
	q.Push(7)
	q.Push(8)
	q.Push(9)
	q.Push(10)

	for i := range 5 {
		val, ok := q.Steal()
		if !ok {
			t.Errorf("Expected %t, but got %t\n", true, ok)
		}

		if val != 10-i {
			t.Errorf("Expected %d, but got %d\n", 10-i, val)
		}
	}

	for i := range 2 {
		val, ok := q.Steal()
		if !ok {
			t.Errorf("Expected %t, but got %t\n", true, ok)
		}

		if val != 2-i {
			t.Errorf("Expected %d, but got %d\n", 2-i, val)
		}
	}
}
