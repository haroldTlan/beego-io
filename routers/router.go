// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"speedio/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/api",
		beego.NSNamespace("/disks",
			beego.NSInclude(
				&controllers.DisksController{},
			),
		),
		beego.NSNamespace("/raids",
			beego.NSInclude(
				&controllers.RaidsController{},
			),
		),
		beego.NSNamespace("/inits",
			beego.NSInclude(
				&controllers.InitsController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
