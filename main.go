package main

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
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
	orm.DefaultTimeLoc = time.Local

	logs.SetLogger(logs.AdapterFile, `{"filename":"/var/log/newSpeedio.log","daily":false,"maxdays":365,"level":3}`)
	logs.EnableFuncCallDepth(true)
	logs.Async()

}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
