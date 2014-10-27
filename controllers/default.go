package controllers

import (
	"github.com/astaxie/beego"
	log "github.com/golang/glog"
	"github.com/trebogeer/timetrap/data"
	gp "github.com/trebogeer/timetrap/graphplot"
	"github.com/trebogeer/timetrap/mongo"
	"github.com/trebogeer/timetrap/simplify"
	//	"sort"
	"bytes"
	"errors"
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

	result func(*TTController, map[string]interface{})
)

const (
	tback_c    = "600s"
	keep_p     = 1200
	dateFormat = "2006-01-02T15:04:05MST"
	split_p    = "12h"
	def_x      = "ts"
	def_y      = "lp"
)

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
		} else {
			this.Ctx.Output.Header("Content-Type", "image/"+ft)
		}
		err := gp.DrawPlot(d, &w, ft)
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
								vis[s] = simplify.Point{float64(v[s][0].(int64)), getFloat64(v[s][1])}
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
			i_b := getInt64T(v[i][0]) < f_millis
			j_b := getInt64T(v[j][0]) >= t_millis
			for i < j && (i_b || j_b) {
				if i_b {
					i = i + 1
					i_b = getInt64T(v[i][0]) < f_millis
				}
				if j_b {
					j = j - 1
					j_b = getInt64T(v[j][0]) >= t_millis
				}
			}
			f_res[k] = v[i:j]
		}
	}
	return f_res
}

func getFloat64(t interface{}) float64 {
	switch t_ := t.(type) {
	case int:
		return float64(t.(int))
	case float64:
		return t.(float64)
	case int64:
		return float64(t.(int64))
	case int32:
		return float64(t.(int32))
	case float32:
		return float64(t.(float32))
	default:
		log.Errorf("Type is unsupported. %v is of type %v.", t, t_)
		return 0

	}
}

func getInt64T(t interface{}) int64 {
	switch t_ := t.(type) {
	case int:
		return int64(t.(int))
	case int64:
		return t.(int64)
	default:
		log.Errorf("Type is unsupported. %v is of type %v.", t, t_)
		return 0
	}
}

func mergeMaps(m *map[string]data.Points, ch *chan map[string]data.Points, ch_cnt int) {
	mm := *m

	for i := 0; i < ch_cnt; i++ {
		m_ := <-*ch
		for k, v := range m_ {
			if val, ok := mm[k]; ok {
				mm[k] = append(val, v...)
			} else {
				mm[k] = v
			}
		}
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
