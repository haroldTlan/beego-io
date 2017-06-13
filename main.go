package main

import (
	"github.com/astaxie/beego"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	_ "speedio/routers"
)

var (
	Dbfile = beego.AppConfig.String("dbfile") //sqlite databases's location
)

func init() {
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", Dbfile)
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
