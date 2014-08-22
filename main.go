package main

import (
	_ "github.com/trebogeer/timetrap/routers"
	"github.com/astaxie/beego"
    "github.com/trebogeer/timetrap/mongo"
)

func main() {
    mongo.Init()
	beego.Run()
}

