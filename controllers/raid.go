package controllers

import (
	"speedio/models"

	"github.com/astaxie/beego"
	"speedio/controllers/web"
)

// Operations about Raids
type RaidsController struct {
	beego.Controller
}

// URLMapping ...
func (c *RaidsController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetAll", c.GetAll)
	//      c.Mapping("Delete", c.Delete)
}

// @Title GetAll
// @Description get all raids
// @Success 200 {object} models.Raids
// @router / [get]
func (c *RaidsController) GetAll() {
	raids, err := models.GetAllRaids()

	result := web.NewResponse(raids, err)
	c.Data["json"] = &result

	c.ServeJSON()
}

// @Title CreateRaid
// @Description create raids
// @Param	body		body 	models.Raid	true		"body for raid content"
// @Success 200 {int} models.Raid.Id
// @Failure 403 body is empty
// @router / [post]
func (c *RaidsController) Post() {
	name := c.GetString("name")
	level, _ := c.GetInt64("level")
	chunk, _ := c.GetInt64("chunk")
	raid := c.GetString("raid_disks")
	spare := c.GetString("spare_disks")

	rebuildPriority := c.GetString("rebuild_priority")
	sync := false

	err := models.AddRaids(name, raid, spare, rebuildPriority, chunk, level, sync)
	result := web.NewResponse(err, err)
	c.Data["json"] = &result

	c.ServeJSON()
}

// @Title Delete
// @Description delete the raid
// @Param	uid		path 	string	true		"The raid you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:name [delete]
func (c *RaidsController) Delete() {
	name := c.Ctx.Input.Param(":name")
	err := models.DelRaids(name)

	result := web.NewResponse(err, err)
	c.Data["json"] = &result
	c.ServeJSON()
}

/*
// @Title Get
// @Description get user by uid
// @Param	uid		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.User
// @Failure 403 :uid is empty
// @router /:uid [get]
func (u *UserController) Get() {
	uid := u.GetString(":uid")
	if uid != "" {
		user, err := models.GetUser(uid)
		if err != nil {
			u.Data["json"] = err.Error()
		} else {
			u.Data["json"] = user
		}
	}
	u.ServeJSON()
}

// @Title Update
// @Description update the user
// @Param	uid		path 	string	true		"The uid you want to update"
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {object} models.User
// @Failure 403 :uid is not int
// @router /:uid [put]
func (u *UserController) Put() {
	uid := u.GetString(":uid")
	if uid != "" {
		var user models.User
		json.Unmarshal(u.Ctx.Input.RequestBody, &user)
		uu, err := models.UpdateUser(uid, &user)
		if err != nil {
			u.Data["json"] = err.Error()
		} else {
			u.Data["json"] = uu
		}
	}
	u.ServeJSON()
}



// @Title Login
// @Description Logs user into the system
// @Param	username		query 	string	true		"The username for login"
// @Param	password		query 	string	true		"The password for login"
// @Success 200 {string} login success
// @Failure 403 user not exist
// @router /login [get]
func (u *UserController) Login() {
	username := u.GetString("username")
	password := u.GetString("password")
	if models.Login(username, password) {
		u.Data["json"] = "login success"
	} else {
		u.Data["json"] = "user not exist"
	}
	u.ServeJSON()
}

// @Title logout
// @Description Logs out current logged in user session
// @Success 200 {string} logout success
// @router /logout [get]
func (u *UserController) Logout() {
	u.Data["json"] = "logout success"
	u.ServeJSON()
}

*/
