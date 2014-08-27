package minheap

import (
	//    "math/rand"
	//    "fmt"
	"errors"
)

type HeapEntry interface {
	SetIndex(i int)
	GetIndex() int
	//Weight(w uint8)
	Weight() float64
}

type MinHeap struct {
	Array []HeapEntry
	Size  int
}

func compare(a HeapEntry, b HeapEntry) int8 {
	if a.Weight() > b.Weight() {
		return 1
	} else if a.Weight() == b.Weight() {
		return 0
	} else {
		return -1
	}
}

func (heap *MinHeap) Push(e HeapEntry) {
	//    if e == nil {
	//       return -1
	//    }
	i := heap.Size
	e.SetIndex(i)
	//e.Weight(uint8(rand.Intn(255)))
	heap.Array[i] = e
	heap.Size = i + 1
	heap.up(i)
}

func (heap *MinHeap) Pop() (error, HeapEntry) {
	if heap.Size == 0 {
		return errors.New("Min Heap is empty"), nil
	}
	// fmt.Printf("Heap.Size = %v\n", heap.Size)
	// fmt.Printf("Array.len = %v\n", len(heap.Array))
	rem := heap.Array[0]
	heap.Size = heap.Size - 1
	if heap.Size > 0 {
		last := heap.Array[heap.Size]
		heap.Array = heap.Array[:heap.Size]
		heap.Array[0] = last
		last.SetIndex(0)
		heap.down(0)
	} else {
		heap.Array = heap.Array[:heap.Size]
	}
	return nil, rem
}

func (heap *MinHeap) down(i int) {
	e := heap.Array[i]
	for {
		right := (i + 1) * 2
		left := right - 1
		down := i
		child := heap.Array[down]
		if left < heap.Size && compare(heap.Array[left], child) < 0 {
			down = left
			child = heap.Array[down]
		}

		if right < heap.Size && compare(heap.Array[right], child) < 0 {
			down = right
			child = heap.Array[down]
		}

		if down == i {
			break
		}
		child.SetIndex(i)
		heap.Array[i] = child
		e.SetIndex(down)
		heap.Array[down] = e
		i = down
	}
}

func (heap *MinHeap) Remove(e HeapEntry) error {
	if heap.Size == 0 {
		return errors.New("Heap is empty. Nothing to remove.")
	}
	i := e.GetIndex()
	last := heap.Array[heap.Size-1]
	heap.Array = heap.Array[:heap.Size-2]
	heap.Size = heap.Size - 1
	if i != heap.Size {
		last.SetIndex(i)
		heap.Array[i] = last
		if compare(last, e) < 0 {
			heap.up(i)
		} else {
			heap.down(i)
		}
	}
	return nil
}

func (heap *MinHeap) up(i int) {
	e := heap.Array[i]
	for i > 0 {
		up := ((i + 1) >> 1) - 1
		parent := heap.Array[up]
		if compare(e, parent) >= 0 {
			break
		}
		parent.SetIndex(i)
		heap.Array[i] = parent
		e.SetIndex(up)
		i = up
	}
}

/*func main() {
    a := HeapEntry{0, 0}
    b := HeapEntry{0, 0}
    c := HeapEntry{0, 0}

    heap := MinHeap{make([]HeapEntry, 3), 0}
    heap.Push(b)
    heap.Push(a)
    heap.Push(c)

    fmt.Printf("Weight: %v\n", heap.Pop().Weight)
    fmt.Printf("len: %v\n", len(heap.Array))
    fmt.Printf("Weight: %v\n", heap.Pop().Weight)
    fmt.Printf("len: %v\n", len(heap.Array))
    fmt.Printf("Weight: %v\n", heap.Pop().Weight)
    fmt.Printf("len: %v\n", len(heap.Array))

}*/
