package data

type (
	XY     [2]interface{}
	Points []XY
)

func (p Points) Len() int {
	return len(p)
}

func (p Points) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Points) Less(i, j int) bool {
	return p[i][0].(int) < p[j][0].(int)
}

func (xy XY) X() interface{} {
	return xy[0]
}

func (xy XY) Y() interface{} {
	return xy[1]
}
