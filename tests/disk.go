package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	_ "github.com/mattn/go-sqlite3"
	"time"
)

func main() {
	o := orm.NewOrm()
	fmt.Println(Dbfile, a)
	var d Disks
	if err := o.QueryTable(new(Disks)).Filter("uuid", "asdasds").One(&d); err != nil {
		fmt.Printf("%+v\n\n!!!!!", err.Error()[0:3])
		fmt.Println(err.Error() == "<QuerySeter> no row found")
	}
	fmt.Printf("%+v", d)
}

var (
	Dbfile = beego.AppConfig.String("dbfile")
	a      = beego.AppConfig.String("runmode")
)

func init() {
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "/root/speediodb.db")
	orm.RegisterModel(new(Disks))
}

type Disks struct {
	Id        string    `orm:"column(uuid);size(255);pk" json:"uuid"`
	Created   time.Time `orm:"column(created_at);type(datetime)"  json:"created_at"`
	Updated   time.Time `orm:"column(updated_at);type(datetime)"  json:"updated_at"`
	Loc       string    `orm:"column(location);size(255)" json:"location"`
	Ploc      string    `orm:"column(prev_location);size(255)" json:"prev_location"`
	Health    string    `orm:"column(health);size(255)" json:"health"`
	Role      string    `orm:"column(role);size(255)" json:"role"`
	Raid      string    `orm:"column(raid);size(255)"  json:"raid"`       //raid's name
	RaidId    string    `orm:"column(raid_id);size(255)"  json:"raid_id"` //raid's uuid
	Disktype  string    `orm:"column(disktype);size(255)" json:"disktype"`
	Vendor    string    `orm:"column(vendor);size(255)" json:"vendor"`
	Model     string    `orm:"column(model);size(255)" json:"model"`
	Host      string    `orm:"column(host);size(255)" json:"host"`
	Sn        string    `orm:"column(sn);size(255)" json:"sn"`
	CapSector int64     `orm:"column(cap_sector)" json:"cap_sector"`
	Link      int64     `orm:"column(link)" json:"link"`
}
