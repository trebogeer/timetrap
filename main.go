package main

import (
	"flag"
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/mongo"
	_ "github.com/trebogeer/timetrap/routers"
	"os"
	"runtime"
)

const (
	vgfpath = "VGFONTPATH"
)

var (
	mongoHost     = flag.String("mhost", "localhost", "Mongo Host")
	mongoPort     = flag.String("mport", "27017", "Mongo Port")
	mongoAuthDB   = flag.String("mauthdb", "admin", "Mongo Auth DB")
	mongoUser     = flag.String("muser", "midori", "Mongo User")
	mongoPassword = flag.String("mpwd", "midori", "Mongo Password")
	beegoPort     = flag.Int("port", 6060, "Timetrap Port")
	runMode       = flag.String("runmode", "prod", "Timetrap Run Mode (dev/prod)")
)

func main() {
	if len(os.Getenv(vgfpath)) == 0 {
		os.Setenv(vgfpath, "./static")
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	mongo.Init(*mongoHost, *mongoPort, *mongoAuthDB, *mongoUser, *mongoPassword)
	beego.HttpPort = *beegoPort
	beego.RunMode = *runMode
	beego.Run()
}
