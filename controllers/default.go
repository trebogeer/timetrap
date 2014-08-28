package controllers

import (
    "log"
    "time"
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/mongo"
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


	dur, err := time.ParseDuration("-"+ tback)
    if err != nil {
        dur,_ = time.ParseDuration("-" + tback_c)
    }
	data := getGraphData(db, x, y, collections, []string{labelName}, time.Now().Add(dur), time.Now())
	/*if err != nil {
		beego.Error(err)
		this.Abort("500")
	}*/

    this.Data["json"] = data
    this.ServeJson()

}


func getGraphData(db, x, y string, collections , labels []string, from, to time.Time) map[string]mongo.Points {
    dur, _ := time.ParseDuration("1h")


    c_len:= len(collections)
    res_channel := make (chan map[string]mongo.Points, 100)
    for i:= 0; i < c_len; i++ {
       go func(c string) {
          t_chan := make(chan map[string]mongo.Points, 1000)
          t := from
          ch_cnt := 0
          for t.Before(to) {
             tt := min(t.Add(dur), to)
             ch_cnt++
             go func(c string, f, t time.Time) {
                err, data := mongo.GetGraphData(db, c, x, y , f, t, labels)
                if err != nil {
                    log.Println(err)
                    t_chan <- make(map[string]mongo.Points)
                } else {
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

    for i:= 0; i < ch_cnt; i++ {
        m_ := <-*ch
        for k,v := range m_ {
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





