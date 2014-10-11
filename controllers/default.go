package controllers

import (
	"github.com/astaxie/beego"
	log "github.com/golang/glog"
	"github.com/trebogeer/timetrap/mongo"
	"github.com/trebogeer/timetrap/simplify"
	"sort"
	"time"
)

type (
	MainController struct {
		beego.Controller
	}

	TTController struct {
		beego.Controller
	}
)

const (
	tback_c = "600s"
	keep_p  = 1200
    dateFormat = "yyyy-MM-dd'T'HH:mm:ssz"
)

func (this *MainController) Get() {
	this.Data["Website"] = "timetrap.io"
	this.Data["Email"] = "trebogeer@gmail.com"
	this.TplNames = "index.tpl"
}

func (this *TTController) GraphData() {
	//&tback=1200&x=lp&label=PRI&labelName=repl&d3=true&labelNameAdd=h&simplify=true&keepPoints=800
	db := this.GetString("db")
	c := this.GetString("c")
	x := this.GetString("x")
    tf := this.GetString("from")


    tt := this.GetString("to")
	if len(x) == 0 {
		x = "ts"
	}
	y := this.GetString("y")
	tback := this.GetString("tback")
	if len(tback) == 0 {
		tback = tback_c
	}
	labelName := this.GetString("labelName")
	keepPoints, err := this.GetInt("keepPoints")
	if err != nil {
		keepPoints = keep_p
	}

	log.V(2).Info(keepPoints)
	if len(c) == 0 || len(db) == 0 {
		beego.Error("Invalid request")
		this.Abort("400")
	}

	err, collections := mongo.GetFilteredCollections(db, c)
	if err != nil {
		log.Error("Failed to get filtered collections")
		beego.Error(err)
		this.Abort("500")
	}

    var f time.Time
    var t time.Time

	dur, err := time.ParseDuration("-" + tback)
	if err != nil {
		dur, _ = time.ParseDuration("-" + tback_c)
	}

    f = time.Now().Add(dur)
    t = time.Now()

    if len(tf) != 0 && len(tt) != 0 {
        if f, err = time.Parse(dateFormat, tf); err != nil {
             log.Error("Failed to parse start date.", tf)
        }
        if t, err = time.Parse(dateFormat, tt); err != nil {
             log.Error("Failed to parse end date.", tt)
        }
    }

	data := getGraphData(db, x, y, collections, []string{labelName}, f, t, keepPoints)
	d := make(map[string]interface{})
	dd := make([]interface{}, 0, len(data))
	d["alias"] = "lp"
	for k, v := range data {
		m := make(map[string]interface{})
		m["key"] = k
		sort.Sort(v)
		m["values"] = v
		dd = append(dd, m)
	}
	d["data"] = dd
	this.Data["json"] = d
	this.ServeJson()

}

func getGraphData(db, x, y string, collections, labels []string, from, to time.Time, to_keep int64) map[string]mongo.Points {
	t_diff := to.Sub(from)
	dur, _ := time.ParseDuration("6h")
	var keep_per_slice int64
	if t_diff > dur {
		slices := t_diff / dur
		keep_per_slice = to_keep / int64(slices)
	} else {
		keep_per_slice = to_keep
	}

	c_len := len(collections)
	res_channel := make(chan map[string]mongo.Points, 100)
	for i := 0; i < c_len; i++ {
		go func(c string) {
			t_chan := make(chan map[string]mongo.Points, 1000)
			t := from
			ch_cnt := 0
			for t.Before(to) {
				tt := min(t.Add(dur), to)
				ch_cnt++
				go func(c string, f, t time.Time) {
					err, data := mongo.GetGraphData(db, c, x, y, f, t, labels)
					if err != nil {
						log.Error(err)
						t_chan <- make(map[string]mongo.Points)
					} else {
						for k, v := range data {
							//sort.Sort(v)
							l := len(v)
							vis := make([]simplify.Point, l)
							for s := 0; s < l; s++ {
								vis[s] = simplify.Point{float64(v[s][0].(int64)), getFloat64(v[s][1])}
							}

							err, viss := simplify.Visvalingam(int(keep_per_slice), vis)
							if err != nil {
								viss = vis
								log.Error("failed to simplify line.", err)
							}
							l = len(viss)
							vv := make(mongo.Points, l)
							for v := 0; v < l; v++ {
								x := int(viss[v].X)
								y := float32(viss[v].Y)
								vv[v] = mongo.XY{x, y}
							}

							data[k] = vv

						}
						t_chan <- data
					}

				}(c, t, tt)
				t = tt
			}
			res := make(map[string]mongo.Points)
			mergeMaps(&res, &t_chan, ch_cnt)
			res_channel <- res
		}(collections[i])
	}
	f_res := make(map[string]mongo.Points)
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
		log.Errorf("Tyoe is unsupported. %v is of type %v.", t, t_)
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
		log.Errorf("Type is unsupprted. %v is of type %v.", t, t_)
		return 0
	}
}

func mergeMaps(m *map[string]mongo.Points, ch *chan map[string]mongo.Points, ch_cnt int) {
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

/*
func (p mongo.Points) Len() int {
    return len(p)
}

func (p mongo.Points) Swap (i, j int) {
    p[i], p[j] = p[j], p[i]
}

func (p mongo.Points) Less(i, j int) {
   return p[i][0].(int) < p[j][0].(int)
}
*/
