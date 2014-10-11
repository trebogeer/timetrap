package main

import (
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/mongo"
	_ "github.com/trebogeer/timetrap/routers"
	"runtime"
    "flag"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
    flag.Parse()
	mongo.Init("greenapi301p.dev.ch3.s.com", "20000", "admin", "midori", "midori")
	beego.HttpPort = 6060
	beego.RunMode = "prod"
	beego.Run()
}
