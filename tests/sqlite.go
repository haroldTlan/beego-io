package main

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Dbfile = beego.AppConfig.String("dbfile") //sqlite databases's location
)

func init() {
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "/root/speediodb.db")
	orm.DefaultTimeLoc = time.Local
	orm.RegisterModel(new(Teachers), new(Students))
}

type Teachers struct {
	Id   int         `orm:"column(Id);pk"	json:"Id"`
	Name int         `orm:"column(Age)"	json:"age"`
	T    []*Students `orm:"reverse(many)"`
}

type Students struct {
	Id  int       `orm:"column(Id);pk"	json:"id"`
	Tid *Teachers `orm:"column(TeacherId);rel(fk)"	json:"TeacherId"`
	//P    `orm:"rel(fk)"`
}

func main() {
	o := orm.NewOrm()
	var t []Teachers
	var s []Students

	o.QueryTable(new(Students)).All(&s)
	fmt.Printf("%+v\n", s)
	for _, ss := range s {
		fmt.Printf("%+v\n", ss.Tid)
		fmt.Printf("%+v\n", ss.Tid.Name)
	}

	o.QueryTable(new(Teachers)).All(&t)
	fmt.Printf("%+v\n", t[0].T)
	//fmt.Printf("%+v\n", t.T)
}
