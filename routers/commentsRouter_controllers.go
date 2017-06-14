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

	beego.GlobalControllerRouter["speedio/controllers:RaidsController"] = append(beego.GlobalControllerRouter["speedio/controllers:RaidsController"],
		beego.ControllerComments{
			Method: "GetAll",
			Router: `/`,
			AllowHTTPMethods: []string{"get"},
			Params: nil})

	beego.GlobalControllerRouter["speedio/controllers:RaidsController"] = append(beego.GlobalControllerRouter["speedio/controllers:RaidsController"],
		beego.ControllerComments{
			Method: "Post",
			Router: `/`,
			AllowHTTPMethods: []string{"post"},
			Params: nil})

	beego.GlobalControllerRouter["speedio/controllers:RaidsController"] = append(beego.GlobalControllerRouter["speedio/controllers:RaidsController"],
		beego.ControllerComments{
			Method: "Delete",
			Router: `/:name`,
			AllowHTTPMethods: []string{"delete"},
			Params: nil})

}
