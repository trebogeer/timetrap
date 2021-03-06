package controllers

import (
	"bytes"
	"errors"
	"github.com/astaxie/beego"
	log "github.com/golang/glog"
	"github.com/trebogeer/timetrap/data"
	gp "github.com/trebogeer/timetrap/graphplot"
	"github.com/trebogeer/timetrap/mongo"
	"github.com/trebogeer/timetrap/simplify"
	"github.com/trebogeer/timetrap/util"
	"sort"
	"time"
	//    "strconv"
)

type (
	MainController struct {
		beego.Controller
	}

	TTController struct {
		beego.Controller
	}

	ReqParams struct {
		db         string
		c          string
		x          string
		y          string
		split      string
		tf         time.Time
		tt         time.Time
		tback      string
		labelName  string
		keepPoints int64
	}

	result          func(*TTController, map[string]interface{})
	SortPointArrays []data.Points
)

const (
	tback_c    = "600s"
	keep_p     = 1200
	dateFormat = "2006-01-02T15:04:05MST"
	split_p    = "12h"
	def_x      = "ts"
	def_y      = "lp"
)

func (spa SortPointArrays) Len() int      { return len(spa) }
func (spa SortPointArrays) Swap(i, j int) { spa[i], spa[j] = spa[j], spa[i] }
func (spa SortPointArrays) Less(i, j int) bool {
	if len(spa[i]) > 0 && len(spa[j]) > 0 {
		return spa[i][0].X().(int) < spa[j][0].X().(int)
	} else {
		return false
	}
}

func (this *MainController) Get() {
	this.Data["Website"] = "timetrap.io"
	this.Data["Email"] = "trebogeer@gmail.com"
	this.TplNames = "index.tpl"
}

func (this *TTController) GraphData(fn result) {
	qp, err := this.parseQueryParams()
	if err != nil {
		log.Error(err)
		beego.Error("Invalid request. Query parameters are either incorrect or missing while are required.")
		this.Abort("400")
	}

	err, collections := mongo.GetFilteredCollections(qp.db, qp.c)
	if err != nil {
		log.Error("Failed to get filtered collections.")
		beego.Error(err)
		this.Abort("500")
	}

	data := getGraphData(qp.db, qp.x, qp.y, qp.split, collections, []string{qp.labelName}, qp.tf, qp.tt, qp.keepPoints)
	d := make(map[string]interface{})
	dd := make([]map[string]interface{}, 0, len(data))
	alias := mongo.GetKV(qp.db, "alias", qp.y)
	d["alias"] = alias

	for k, v := range data {
		m := make(map[string]interface{})
		m["key"] = k
		//sort.Sort(v)
		m["values"] = v
		dd = append(dd, m)
	}
	d["data"] = dd
	fn(this, d)
}

func (this *TTController) GraphDataImage() {
	this.GraphData(func(this *TTController, d map[string]interface{}) {
		var w bytes.Buffer
		ft := this.GetString("ft")
		if len(ft) == 0 {
			ft = this.Ctx.Input.Header("Accept")
			if len(ft) == 0 {
				ft = "png"
			}
		}

		if ft == "pdf" {
			this.Ctx.Output.Header("Content-Type", "application/pdf")
		} else if ft == "svg" {
			this.Ctx.Output.Header("Content-Type", "image/svg+xml")
		} else {
			this.Ctx.Output.Header("Content-Type", "image/"+ft)
		}

		width, err := this.GetFloat("w")
		if err != nil {
			width = 0
		}

		height, err := this.GetFloat("h")
		if err != nil {
			height = 0
		}

		showLegend, err := this.GetBool("shl")
		if err != nil {
			showLegend = true
		}

		err = gp.DrawPlot(d, &w, ft, width, height, showLegend)
		log.V(1).Info("Draw.", w.Len())
		if err != nil {
			log.Error("Failed to write image to response writer.")
			beego.Error(err)
			this.Abort("500")
		}
		this.Ctx.Output.Body(w.Bytes())
		//this.Data["json"] = make([]int, 1)
		//this.ServeJson()
	})
}

func (this *TTController) GraphDataJson() {
	this.GraphData(func(this *TTController, d map[string]interface{}) {
		this.Data["json"] = d
		this.ServeJson()
	})
}

func getGraphData(db, x, y, split string, collections, labels []string, from, to time.Time, to_keep int64) map[string]data.Points {
	t_diff := to.Sub(from)
	dur, err := time.ParseDuration(split)
	if err != nil {
		dur = t_diff
	}
	dur = minDur(t_diff, dur)
	var keep_per_slice int64
	if t_diff > dur {
		slices := t_diff / dur
		keep_per_slice = to_keep / int64(slices)
	} else {
		keep_per_slice = to_keep
	}

	c_len := len(collections)
	res_channel := make(chan map[string]data.Points, 100)
	for i := 0; i < c_len; i++ {
		go func(c string) {
			t_chan := make(chan map[string]data.Points, 1000)
			t := from
			ch_cnt := 0
			for t.Before(to) {
				tt := min(t.Add(dur), to)
				ch_cnt++
				go func(c string, f, t time.Time) {
					err, data_ := mongo.GetGraphData(db, c, x, y, f, t, labels)
					if err != nil {
						log.Error(err)
						t_chan <- make(map[string]data.Points)
					} else {
						for k, v := range data_ {
							l := len(v)
							vis := make([]simplify.Point, l)
							for s := 0; s < l; s++ {
								vis[s] = simplify.Point{util.AssertFloat64(v[s][0], 0), util.AssertFloat64(v[s][1], 0)}
							}

							err, viss := simplify.Visvalingam(int(keep_per_slice), vis)
							if err != nil {
								viss = vis
								log.V(1).Info("Failed to simplify line. ", err)
							}
							l = len(viss)
							vv := make(data.Points, l)
							for v := 0; v < l; v++ {
								x := int(viss[v].X)
								y := float32(viss[v].Y)
								vv[v] = data.XY{x, y}
							}

							data_[k] = vv

						}
						t_chan <- data_
					}

				}(c, t, tt)
				t = tt
			}
			res := make(map[string]data.Points)
			mergeMaps(&res, &t_chan, ch_cnt)
			res_channel <- res
		}(collections[i])
	}
	f_res := make(map[string]data.Points)
	mergeMaps(&f_res, &res_channel, c_len)

	f_millis := from.UnixNano() / int64(time.Millisecond)
	t_millis := to.UnixNano() / int64(time.Millisecond)
	for k, v := range f_res {
		if len(v) > 1 {
			i := 0
			j := len(v) - 1
			i_b := util.AssertInt64(v[i][0], 0) < f_millis
			j_b := util.AssertInt64(v[j][0], 0) >= t_millis
			for i < j && (i_b || j_b) {
				if i_b {
					i = i + 1
					i_b = util.AssertInt64(v[i][0], 0) < f_millis
				}
				if j_b {
					j = j - 1
					j_b = util.AssertInt64(v[j][0], 0) >= t_millis
				}
			}
			f_res[k] = v[i:j]
		}
	}
	return f_res
}

// merge preserving point order. Looking at first elements of
// point arrays assuming they are sorted already.
func mergeMaps(m *map[string]data.Points, ch *chan map[string]data.Points, ch_cnt int) {
	mm := *m
	mm_ := make(map[string]SortPointArrays)
	for i := 0; i < ch_cnt; i++ {
		m_ := <-*ch
		for k, v := range m_ {
			if val, ok := mm_[k]; ok {
				mm_[k] = append(val, v)
				log.V(2).Infof("M[K] %v", mm_[k])
			} else {
				a := SortPointArrays([]data.Points{v})
				mm_[k] = a
				log.V(2).Infof("M[1] %v", mm_[k])
			}
		}
	}

	log.V(2).Infof("Merge: %v", mm_)

	for k, v := range mm_ {
		var arr data.Points
		if len(v) > 0 {
			sort.Sort(v)
			for i := 0; i < len(v); i++ {
				arr = append(arr, v[i]...)
			}
		}
		mm[k] = arr
	}
}

func min(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	} else {
		return b
	}
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	} else {
		return b
	}
}

func (this *TTController) parseQueryParams() (ReqParams, error) {
	qp := ReqParams{}
	//&tback=1200&x=lp&label=PRI&labelName=repl&d3=true&labelNameAdd=h&simplify=true&keepPoints=800
	qp.db = this.GetString("db")
	qp.c = this.GetString("c")

	if len(qp.c) == 0 || len(qp.db) == 0 {
		return qp, errors.New("'db' and 'c' query parameters must be present.")
	}

	qp.x = this.GetString("x")
	qp.split = this.GetString("split")
	if len(qp.split) == 0 {
		qp.split = split_p
	}

	qp.tback = this.GetString("tback")
	if len(qp.tback) == 0 {
		qp.tback = tback_c
	}

	tf := this.GetString("from")
	tt := this.GetString("to")

	var f time.Time
	var t time.Time

	dur, err := time.ParseDuration("-" + qp.tback)
	if err != nil {
		dur, _ = time.ParseDuration("-" + tback_c)
	}

	f = time.Now().Add(dur)
	t = time.Now()

	if len(tf) != 0 && len(tt) != 0 {
		f_ := f
		if f, err = time.Parse(dateFormat, tf); err != nil {
			log.Error("Failed to parse start date. Falling back to defaults.", tf)
		} else if t, err = time.Parse(dateFormat, tt); err != nil {
			log.Error("Failed to parse end date. Falling back to defaults.", tt)
			f = f_
		}
	}

	qp.tf = f
	qp.tt = t

	if len(qp.x) == 0 {
		qp.x = def_x
	}
	qp.y = this.GetString("y")
	if len(qp.y) == 0 {
		qp.y = def_y
	}

	qp.labelName = this.GetString("labelName")
	qp.keepPoints, err = this.GetInt("keepPoints")
	if err != nil {
		qp.keepPoints = keep_p
	}
	return qp, nil
}
