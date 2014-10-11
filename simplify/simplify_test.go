package simplify

import (
	"testing"
	//"fmt"
)

func TestVisvalingam(t *testing.T) {
	p := []Point{{4, 7}, {6, 4}, {8, 6}, {10, 3}, {12, 8}, {16, 9}, {18, 10}, {19, 12}}
	err, p2 := Visvalingam(4, p)
	if err != nil {
		t.Error(err.Error())
	}
	if len(p2) != 4 {
		t.Error("Failed!")
		t.Errorf("Current length %v", len(p2))
	}
}

func TestVisvalingamSort(t *testing.T) {
	p := []Point{{5, 6}, {3, 7}, {2, 5}, {1, 8}, {7, 12}, {8, 34}}
	err, p2 := Visvalingam(4, p)
	if err != nil {
		t.Error(err.Error())
	}

	for i := range p2 {
		if i != 0 && p2[i-1].X > p2[i].X {
			t.Error("Order is incorrect. Failed!")
			t.Errorf("X in %v is %v\n", i, p2[i].X)
		}
	}
}

func TestVisvalingamCC(t *testing.T) {

	p := []Point{}

	err, _ := Visvalingam(2, p)

	if err == nil {
		t.Error("Empty array error is expected but not reported")
	}

	p = []Point{{4, 5}}

	err, r := Visvalingam(1, p)

	if err != nil {
		t.Error(err.Error())
	}

	if r[0].X != 4 {
		t.Error("Point has wrong value for X")
	}

	p = []Point{{3, 5}, {2, 6}}

	err, r = Visvalingam(2, p)

	if err != nil {
		t.Error(err.Error())
	}

	if len(r) != 2 {
		t.Error("Incorrect simplification. Result array size is incorrect.")
	}

}
