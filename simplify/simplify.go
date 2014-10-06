package simplify

import (
	//"fmt"
	"errors"
	mh "github.com/trebogeer/timetrap/minheap"
	"math"
	"sort"
	//    "container/list"
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

/*func (t1 Triangle) Compare(t2 Triangle) int8 {
    if t1.Area > t2.Area {
        return 1
    } else if t1.Area == t2.Area {
        return 0//mh.HeapEntry{t1.Index, t1.Weight}.Compare(mh.HeapEntry{t2.Index, t2.Weight})
    } else {
        return -1
    }
}*/
/*
Sorting interface implementation
*/
type ByX []Point

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

func (t Triangle) Weight() float64 {
	return t.Area
}

func (t Triangle) GetIndex() int {
	return t.ind
}

func (t Triangle) SetIndex(i int) {
	t.ind = i
}

func (t *Triangle) TArea() float64 {
	area := area(*t)
	t.Area = area
	return area
}

func area(t Triangle) float64 {
	a := t.Prev
	b := t.Point
	c := t.Next
	return math.Abs((a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y)) / 2.0)
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

	tl := mh.MinHeap{make([]mh.HeapEntry, points_len), 0}
	l := points_len - 1
	var tt *Triangle
	for i := 1; i < l; i++ {
		t := Triangle{}
		t.Prev = points[i-1]
		t.Point = points[i]
		t.Next = points[i+1]
		t.Area = area(t)
		if t.Area > 0 {
			tl.Push(t)
			if tt != nil {
				tt.NextT = &t
				t.PrevT = tt
			}

		}
		tt = &t
	}
    cnt := 0
	if keepPoints < tl.Size {
		for tl.Size > keepPoints {
			err, e := tl.Pop()
			if err != nil {
				break
			}
            cnt++
			t := e.(Triangle)
//            fmt.Printf("%v POP: %v\n",cnt, t.Area)
			if t.PrevT != nil {
				t.PrevT.NextT = t.NextT
				t.PrevT.Next = t.Next
				updateHeap(t.PrevT, &tl, t)
			}

			if t.NextT != nil {
				t.NextT.PrevT = t.PrevT
				t.NextT.Prev = t.Prev
				updateHeap(t.NextT, &tl, t)
			}
		}
	}
	res := make([]Point, tl.Size+2)
	tr_res := tl.Array[:tl.Size]
	res[0] = points[0]
	for i := range tr_res {
		t := tr_res[i].(Triangle)
		res[i+1] = t.Point
	}
	res[tl.Size + 1] = points[points_len-1]
	sort.Sort(ByX(res))
	return nil, res

}

func updateHeap(prev *Triangle, heap *mh.MinHeap, t Triangle) {
	err := heap.Remove(prev)
	if err == nil {
		prev.Area = math.Max(area(*prev), t.Area)
		heap.Push(prev)
	}
}
