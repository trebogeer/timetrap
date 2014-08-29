package routers

import (
	"github.com/astaxie/beego"
	"github.com/trebogeer/timetrap/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/data.json", new(controllers.TTController), "get:GraphData")
}
