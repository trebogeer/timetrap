package graphplot

//package main

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	log "github.com/golang/glog"
	d "github.com/trebogeer/timetrap/data"
	"github.com/trebogeer/timetrap/util"
	"strconv"
	"time"
)

const (
	dateFormat = "HH:mm:ss MM/dd/yyyy"
)

func DrawPlot(input map[string]interface{}) error {

	p, err := plot.New()
	if err != nil {
		log.Error("Failed to initialize plot.", err)
		return err
		//panic(err)
	}
	log.Info("Created plot.")
	name := util.AssertString(input["alias"], "N/A")
	p.Title.Text = name
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	log.Info("Set name complete.")
	// Use a custom tick marker function that computes the default
	// tick marks and re-labels the major ticks with dates
	p.X.Tick.Marker = dateTicks
	data := input["data"].([]map[string]interface{})
	log.Info("Initialied data")
	//lines := make([]interface{}, 2*len(data))
	for i := range data {
		//  l := 2 * i
		//lines[l] = data[i]["key"]
		//lines[l + 1] =  makePoints(data[i]["values"].(d.Points))
		err = plotutil.AddLinePoints(p, data[i]["key"], makePoints(data[i]["values"].(d.Points)))
		if err != nil {
			log.Error(err)
		}
	}
	log.Info("Created points.")
	/* for i:= range lines {
	      log.Info(lines[i])
	      plotutil.AddLinePoints(p, data)
	    }
		err = plotutil.AddLinePoints(p, lines)
		if err != nil {
	        log.Error(err)
			return err
		}*/
	log.Info("Added points to plot.")

	// Save the plot to a PNG file.
	if err := p.Save(10, 6, "/tmp/tt.png"); err != nil {
		return err
	}
	log.Info("Saved to file.")
	return nil
}

// RandomPoints returns some random x, y points.
func makePoints(arr d.Points) plotter.XYs {
	pts := make(plotter.XYs, len(arr))
	for i := range pts {
		pts[i].X = util.AssertFloat64(arr[i].X(), 0)
		pts[i].Y = util.AssertFloat64(arr[i].Y(), 0)
	}
	return pts
}

// CommaTicks computes the default tick marks, but inserts commas
// into the labels for the major tick marks.
func dateTicks(min, max float64) []plot.Tick {
	tks := plot.DefaultTicks(min, max)
	log.Info("Format labels.")
	for i, t := range tks {
		if t.Label == "" { // Skip minor ticks, they are fine.
			continue
		}
		tks[i].Label = formatLabel(t.Label)
	}
	log.Info("Format labels done.")
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
	label := t.Format(dateFormat)
	log.Info(label)
	return label
}
