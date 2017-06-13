package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["speedio/controllers:DisksController"] = append(beego.GlobalControllerRouter["speedio/controllers:DisksController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["speedio/controllers:DisksController"] = append(beego.GlobalControllerRouter["speedio/controllers:DisksController"],
		beego.ControllerComments{
			Method: "Format",
			Router: `/:loc`,
			AllowHTTPMethods: []string{"put"},
			Params: nil})

}
