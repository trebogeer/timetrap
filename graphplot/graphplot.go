package graphplot

import (
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgeps"
	"code.google.com/p/plotinum/vg/vgimg"
	"code.google.com/p/plotinum/vg/vgpdf"
	"code.google.com/p/plotinum/vg/vgsvg"
	"errors"
	log "github.com/golang/glog"
	d "github.com/trebogeer/timetrap/data"
	"github.com/trebogeer/timetrap/util"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	def_width  = 12
	def_height = 6
	dateFormat = "15:04:05 01/02/2006" //"HH:mm:ss MM/dd/yyyy"
)

func DrawPlot(input map[string]interface{}, writer io.Writer, ft string) error {

	p, err := plot.New()
	if err != nil {
		log.Error("Failed to initialize plot.", err)
		return err
		//panic(err)
	}
	log.V(2).Info("Created plot.")
	name := util.AssertString(input["alias"], "N/A")
	p.Title.Text = name
	//p.X.Label.Text = "X"
	//p.Y.Label.Text = "Y"
	log.V(2).Info("Set name complete.")
	// Use a custom tick marker function that computes the default
	// tick marks and re-labels the major ticks with dates
	p.X.Tick.Marker = dateTicks
	data := input["data"].([]map[string]interface{})
	for i := range data {
        points := makePoints(data[i]["values"].(d.Points))
        log.V(2).Info("Points length: ", len(points))
		err = AddLinePoints(p, util.AssertString(data[i]["key"], ""), points, i)
		if err != nil {
			log.Error(err)
		}
	}
	log.V(2).Info("Created points.")
	// Save the plot to a PNG file.
	if err := save(p, def_width, def_height, writer, ft); err != nil {
		return err
	}
	log.V(2).Info("Done writing to output.")
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

// dateTicks computes date ticks in format HH:mm:ss MM/dd/yyyy.
func dateTicks(min, max float64) []plot.Tick {
	tks := plot.DefaultTicks(min, max)
	log.V(2).Info("Format labels.")
	for i, t := range tks {
		if t.Label == "" { // Skip minor ticks, they are fine.
			continue
		}
		tks[i].Label = formatLabel(t.Label)
	}
	log.V(2).Info("Format labels done.")
	return tks
}

func formatLabel(s string) string {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.V(2).Info("Failed to convert " + s)
		return s
	}
	i64 := int64(i)
	sec := i64 / 1000
	nanos := (i64 % 1000) * 1000 * 1000

	t := time.Unix(sec, nanos)
	label := t.Format(dateFormat)
	log.V(2).Info("Label : " + label)
	return label
}

func save(p *plot.Plot, width, height float64, writer io.Writer, ft string) (err error) {
	file := ft
	w, h := vg.Inches(width), vg.Inches(height)
	var c interface {
		vg.Canvas
		Size() (w, h vg.Length)
		io.WriterTo
	}
	switch ext := strings.ToLower(file); ext {

	case "eps":
		c = vgeps.NewTitle(w, h, file)

	case "jpg", "jpeg":
		c = vgimg.JpegCanvas{Canvas: vgimg.New(w, h)}

	case "pdf":
		c = vgpdf.New(w, h)

	case "png":
		c = vgimg.PngCanvas{Canvas: vgimg.New(w, h)}

	case "svg":
		c = vgsvg.New(w, h)

	case "tiff":
		c = vgimg.TiffCanvas{Canvas: vgimg.New(w, h)}

	default:
		return errors.New("Unsupported file extension: " + ext)
	}
	p.Draw(plot.MakeDrawArea(c))
	_, err = c.WriteTo(writer)
	return err
}

func AddLinePoints(plt *plot.Plot, name string, points plotter.XYs, i int) error {

	l, s, err := plotter.NewLinePoints(points)
	if err != nil {
		return err
	}
	l.Color = plotutil.Color(i)
	l.Dashes = plotutil.Dashes(i)
	s.Color = plotutil.Color(i)
	s.Shape = plotutil.Shape(i)

	plt.Add(l, s)
	plt.Legend.Add(name, l, s)
	return nil
}
