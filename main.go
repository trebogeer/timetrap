package main

import (
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/mongo"
	_ "github.com/trebogeer/timetrap/routers"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	mongo.Init("greenapi301p.dev.ch3.s.com", "20000", "admin", "midori", "midori")
	beego.HttpPort = 6060
	beego.RunMode = "prod"
	beego.Run()
}
