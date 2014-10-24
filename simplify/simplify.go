package simplify

import (
	"container/heap"
	"errors"
	"math"
	"sort"
)

type Point struct {
	X, Y float64
}

type Triangle struct {
	ind   int
	Next  Point
	Prev  Point
	Point Point
	Area  float64
	NextT *Triangle
	PrevT *Triangle
}

/*
Sorting interface implementation
*/
type ByX []Point

type MinHeap []*Triangle

func (b ByX) Len() int {
	return len(b)
}

func (b ByX) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByX) Less(i, j int) bool {
	return b[i].X < b[j].X
}

/*
MinHeap implementation
*/

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].Area < h[j].Area }
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].ind = i
	h[j].ind = j
}

func (h *MinHeap) Push(x interface{}) {
	i := len(*h)
	entry := x.(*Triangle)
	entry.ind = i
	*h = append(*h, entry)
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	i := len(old)
	entry := old[i-1]
	entry.ind = -1
	*h = old[0 : i-1]
	return entry
}

func area(t Triangle) float64 {
	a := t.Prev
	b := t.Point
	c := t.Next
	return math.Abs((a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y)) / 2)
}

func Visvalingam(toKeep int, points []Point) (error, []Point) {

	points_len := len(points)

	if points_len == 0 {
		return errors.New("Points is empty"), points
	}

	if toKeep > points_len {
		return errors.New("Points array is less then number to keep."), points
	}

	if toKeep == 0 {
		return errors.New("Cannot keep 0 points."), points
	}

	if toKeep == points_len {
		return nil, points
	}

	if toKeep == 1 {
		return nil, points[0:1]
	}

	sort.Sort(ByX(points))

	if toKeep == 2 {
		return nil, append(points[0:1], points[points_len-1])
	}
	keepPoints := toKeep - 2

	tl := make(MinHeap, points_len-2)
	l := points_len - 1
	for i := 1; i < l; i++ {
		t := Triangle{}
		index := i - 1
		t.Prev = points[index]
		t.Point = points[i]
		t.Next = points[i+1]
		t.Area = area(t)
		t.ind = index
		tl[index] = &t
		if index > 0 {
			tl[i-2].NextT = tl[index]
			tl[index].PrevT = tl[i-2]
		}
	}
	heap.Init(&tl)
	if keepPoints < tl.Len() {
		for len(tl) > keepPoints {
			e := heap.Pop(&tl)

			t := e.(*Triangle)
			// log.Printf("Area : %v", t.Area)
			if t.PrevT != nil {
				t.PrevT.NextT = t.NextT
				t.PrevT.Next = t.Next
				tl.updateHeap(t.PrevT, t)
			}

			if t.NextT != nil {
				t.NextT.PrevT = t.PrevT
				t.NextT.Prev = t.Prev
				tl.updateHeap(t.NextT, t)
			}
		}
	}
	res := make([]Point, tl.Len()+2)
	res[0] = points[0]
	for i := range tl {
		t := tl[i]
		res[i+1] = t.Point
	}
	res[tl.Len()+1] = points[points_len-1]
	sort.Sort(ByX(res))
	return nil, res

}

func (h *MinHeap) updateHeap(toFix *Triangle, t *Triangle) {
	toFix.Area = math.Max(area(*toFix), t.Area)
	heap.Fix(h, toFix.ind)
}
