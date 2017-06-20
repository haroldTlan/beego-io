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
	Id       int         `orm:"column(Id);pk"	json:"id"`
	Name     int         `orm:"column(Age)"	json:"age"`
	Students []*Students `orm:"column(TeacherId);reverse(many)"`
}

type Students struct {
	Id  int       `orm:"column(Id);pk"	json:"id"`
	Tid *Teachers `orm:"column(TeacherId);rel(fk)"	json:"teacherId"`
	//	P   *Teachers `orm:"column(Id);rel(fk)"`
}

func main() {
	o := orm.NewOrm()

	t := Teachers{Id: 1}
	err := o.Read(&t)

	fmt.Printf("%+v\n\n", t, err)

	num, err := o.LoadRelated(&t, "Students")
	fmt.Printf("%+v\n\n", num, err)

	fmt.Printf("%+v\n\n", t)
	/*var t Teachers
	var s []*Students

	o.QueryTable(new(Students)).RelatedSel().All(&s)
	fmt.Printf("%+v\n\n", s)
	for _, ss := range s {
		fmt.Printf("%+v\n", ss.Tid)
		fmt.Printf("%+v\n", ss.Tid.Name)
	}

	err := o.QueryTable(new(Teachers)).RelatedSel().One(&t)
	//err := o.QueryTable(new(Teachers)).Filter("Students__TeacherId", "1").Limit(1).One(&t)
	fmt.Printf("%+v\n", t, err)*/
	//fmt.Printf("%+v\n", t[0].T)
	//fmt.Printf("%+v\n", t.T)
}
