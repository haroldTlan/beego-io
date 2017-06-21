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
	orm.RegisterModel(new(T), new(S))
}

type T struct {
	Id   int  `orm:"column(Id);pk"	json:"id"`
	Name int  `orm:"column(Age)"	json:"age"`
	Ss   []*S `orm:"column(TeacherId);reverse(many)"`
}

type S struct {
	Id  int `orm:"column(Id);pk"	json:"id"`
	Tid *T  `orm:"column(TeacherId);rel(fk)"	json:"teacherId"`
	//	P   *Teachers `orm:"column(Id);rel(fk)"`
}

func (d *S) TableName() string {
	return "Students"
}

func (d *T) TableName() string {
	return "Teachers"
}

func main() {
	o := orm.NewOrm()

	t := T{Id: 1}
	o.Read(&t)

	o.LoadRelated(&t, "Ss")
	//fmt.Printf("%+v\n\n", num, err)

	fmt.Printf("%+v\n\n", t)

	for _, i := range t.Ss {
		fmt.Printf("%+v\n\n", i.Tid.Name)
		fmt.Printf("%+v\n\n", *i)
	}

	//exist := o.QueryTable(new(Students)).Filter("teacherId", 1).Exist()
	//fmt.Printf("%+v\n\n", exist)
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
