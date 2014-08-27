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
	tback_c = 600
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
	tback, err := this.GetInt("tback")
	if err != nil {
		tback = tback_c
	}
/*	labelName := this.GetString("labelName")
	aLabel := this.GetString("alabel")
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


	dur, _ := time.ParseDuration("-1000h")
	err, data := mongo.GetGraphData(db, collections[0], x, y, time.Now().Add(dur), time.Now(), []string{labelName})
	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}

    this.Data["json"] = data
    this.ServeJson()

}

