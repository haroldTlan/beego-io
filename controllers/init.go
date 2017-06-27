package controllers

import (
	_ "encoding/json"
	"github.com/astaxie/beego"
	"speedio/controllers/web"
	"speedio/models"
)

// Operations about inits
type InitsController struct {
	beego.Controller
}

// URLMapping ...
func (c *InitsController) URLMapping() {
	c.Mapping("Post", c.Post)
}

// @Title init
// @Description init all
// @Param	body		body 	models.Inits	true		"The inits content"
// @Success 200 {success}
// @Failure 403 body is empty
// @router / [get]
func (c *InitsController) Post() {

	res, err := models.Inits()
	result := web.NewResponse(res, err)
	c.Data["json"] = result
	c.ServeJSON()
}
