package graphplot

//package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	log "github.com/golang/glog"
	"strconv"
	"time"
)

const (
	dateFormat = "HH:mm:ss MM/dd/yyyy"
)

func DrawPlot(name string, data []map[string]interface{}) {

	p, err := plot.New()
	if err != nil {
		log.Error("Failed to initialize plot.", err)
		return
		//panic(err)
	}

	p.Title.Text = name
	p.X.Label.Text = ""
	p.Y.Label.Text = ""
	// Use a custom tick marker function that computes the default
	// tick marks and re-labels the major ticks with dates
	p.X.Tick.Marker = dateTicks

	lines := make([]interface{}, 2*len(data))
	for i := range data {
		l := 2 * i
		lines[l] = data[i]["key"]
		lines[l+1] = makePoints(data[i]["values"].([]interface{}))
	}

	err = plotutil.AddLinePoints(p, lines)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(4, 4, "points_commas.png"); err != nil {
		panic(err)
	}
}

// RandomPoints returns some random x, y points.
func makePoints(arr []interface{}) plotter.XYs {
	pts := make(plotter.XYs, len(arr))
	/*for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = (pts[i].X + 10*rand.Float64()) * 1000
	}*/
	return pts
}

// CommaTicks computes the default tick marks, but inserts commas
// into the labels for the major tick marks.
func dateTicks(min, max float64) []plot.Tick {
	tks := plot.DefaultTicks(min, max)
	for i, t := range tks {
		if t.Label == "" { // Skip minor ticks, they are fine.
			continue
		}
		tks[i].Label = formatLabel(t.Label)
	}
	return tks
}

func formatLabel(s string) string {
	i, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	i64 := int64(i)
	sec := i64 / 1000
	nanos := (i64 % 1000) * 1000 * 1000

	t := time.Unix(sec, nanos)

	return t.Format(dateFormat)
}
