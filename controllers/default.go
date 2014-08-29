package controllers

import (
	"log"
	"time"
	//    "sort"
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/mongo"
	"github.com/trebogeer/timetrap/visvalingam"
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
)

func (this *MainController) Get() {
	this.Data["Website"] = "beego.me"
	this.Data["Email"] = "astaxie@gmail.com"
	this.TplNames = "index.tpl"
}

func (this *TTController) GraphData() {
	//&tback=1200&x=lp&label=PRI&labelName=repl&d3=true&labelNameAdd=h&simplify=true&keepPoints=800
	db := this.GetString("db")
	c := this.GetString("c")
	x := this.GetString("x")
	if len(x) == 0 {
		x = "ts"
	}
	y := this.GetString("y")
	tback := this.GetString("tback")
	if len(tback) == 0 {
		tback = tback_c
	}
	labelName := this.GetString("labelName")
	/*	aLabel := this.GetString("alabel")
		simplify, err := this.GetBool("simplify")
	    if err != nil {
	      simplify = false
	    }

		log.Println(tback)
		log.Println(aLabel)
		log.Println(simplify)
	*/
	keepPoints, err := this.GetInt("keepPoints")
	if err != nil {
		keepPoints = keep_p
	}

	log.Println(keepPoints)
	if len(c) == 0 || len(db) == 0 {
		beego.Error("Invalid request")
		this.Abort("400")
	}

	err, collections := mongo.GetFilteredCollections(db, c)
	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}

	dur, err := time.ParseDuration("-" + tback)
	if err != nil {
		dur, _ = time.ParseDuration("-" + tback_c)
	}
	data := getGraphData(db, x, y, collections, []string{labelName}, time.Now().Add(dur), time.Now(), keepPoints)
	/*if err != nil {
		beego.Error(err)
		this.Abort("500")
	}*/

	this.Data["json"] = data
	this.ServeJson()

}

func getGraphData(db, x, y string, collections, labels []string, from, to time.Time, to_keep int64) map[string]mongo.Points {
    t_diff := to.Sub(from)
	dur, _ := time.ParseDuration("6h")
    slices := t_diff/dur
    keep_per_slice := to_keep/int64(slices)

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
						log.Println(err)
						t_chan <- make(map[string]mongo.Points)
					} else {
						for k, v := range data {
							//sort.Sort(v)
							l := len(v)
							vis := make([]visvalingam.Point, l)
							for s := 0; s < l; s++ {
								vis[s] = visvalingam.Point{float64(v[s][0].(int64)), v[s][1].(float64)}
							}

							err, viss := visvalingam.Visvalingam(int(keep_per_slice), vis)
							if err != nil {
								viss = vis
								log.Println("failed to simplify line.")
								log.Println(err)
							}
							l = len(viss)
							vv := make(mongo.Points, l)
							for v := 0; v < l; v++ {
                                x := int(viss[v].X)
                                y := float32(viss[v].Y)
								vv[v] = mongo.XY{x, y}
   //                             log.Printf("I: %v, X: %v, Y: %v", v, x, y)
							}

							data[k] = vv

						}
						// TODO visvalingam
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
	return f_res
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
