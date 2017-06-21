package controllers

import (
	_ "encoding/json"
	"github.com/astaxie/beego"
	"speedio/controllers/web"
	"speedio/models"
)

// Operations about disks
type DisksController struct {
	beego.Controller
}

// URLMapping ...
func (c *DisksController) URLMapping() {
	// c.Mapping("Post", c.Post)
	c.Mapping("GetAll", c.GetAll)
	//	    c.Mapping("Delete", c.Delete)
}

// @Title GetAll
// @Description get all disks
// @Success 200 {object} models.Disks
// @Failure 403
// @router / [get]
func (c *DisksController) GetAll() {
	disks, err := models.GetRestDisks()

	result := web.NewResponse(disks, err)
	c.Data["json"] = &result
	c.ServeJSON()
}

// @Title Format
// @Description format disk
// @Success 200 {object} models.FormatDisk
// @Failure 403
// @router /:loc [put]
func (c *DisksController) Format() {
	loc := c.Ctx.Input.Param(":loc")

	res, err := models.FormatDisks(loc)

	result := web.NewResponse(res, err)
	c.Data["json"] = &result
	c.ServeJSON()
}

/*
// @Title Create
// @Description create disks
// @Param	body		body 	models.Disks	true		"The disks content"
// @Success 200 {string} models.Disks.Uuid
// @Failure 403 body is empty
// @router / [post]
func (o *DisksController) Post() {
	var ob models.Disks
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)
	objectid := models.AddOne(ob)
	o.Data["json"] = map[string]string{"ObjectId": objectid}
	o.ServeJSON()
}




// @Title Update
// @Description update the object
// @Param	objectId		path 	string	true		"The objectid you want to update"
// @Param	body		body 	models.Object	true		"The body"
// @Success 200 {object} models.Object
// @Failure 403 :objectId is empty
// @router /:objectId [put]
func (o *ObjectController) Put() {
	objectId := o.Ctx.Input.Param(":objectId")
	var ob models.Object
	json.Unmarshal(o.Ctx.Input.RequestBody, &ob)

	err := models.Update(objectId, ob.Score)
	if err != nil {
		o.Data["json"] = err.Error()
	} else {
		o.Data["json"] = "update success!"
	}
	o.ServeJSON()
}

// @Title Delete
// @Description delete the object
// @Param	objectId		path 	string	true		"The objectId you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 objectId is empty
// @router /:objectId [delete]
func (o *DisksController) Delete() {
	objectId := o.Ctx.Input.Param(":objectId")
	models.Delete(objectId)
	o.Data["json"] = "delete success!"
	o.ServeJSON()
}
*/
