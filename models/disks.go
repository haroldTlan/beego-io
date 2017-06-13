package models

import (
	/*"errors"
	  "strconv"
	*/
	"fmt"
	"github.com/astaxie/beego/orm"
	"speedio/models/util"
	"strings"
	"time"
)

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

var (
	DISKTYPE_SATA = "sata"
	DISKTYPE_SAS  = "sas"
	DISKTYPE_SSD  = "ssd"

	HOST_NATIVE  = "native"
	HOST_LOCAL   = "local" //db has no this disk's data, but metadata host uuid equal with storage
	HOST_FOREIGN = "foreign"
	HOST_USED    = "used"

	//						    HEALTH = ["failed", "normal"]

	//							    ROLE = ["unused", "data", "spare", "global spare"]

)

func init() {
	orm.RegisterModel(new(Disks))
}

// GET
// Get all disks
func GetAllDisks() (ds []Disks, err error) {
	o := orm.NewOrm()

	if _, err = o.QueryTable(new(Disks)).All(&ds); err != nil {
		fmt.Printf("%+v\n\n!!!!!", err)
		//TODO 	        AddLog(err)
	}

	return
}

// PUT
// Format disks
func FormatDisks(locations string) (res string, err error) {
	// when loc == all
	// format all
	if locations == "all" {
		disks, err := GetAllDisks()
		if err != nil {
			//TODO 	        AddLog(err)
		}
		for _, disk := range disks {
			if disk.Host == HOST_USED || disk.Host == HOST_FOREIGN {
				err = disk.format()
				//TODO 	        AddLog(err)

			}
		}

		// format single disk
	} else {
		locs := strings.Split(locations, ",")
		for _, loc := range locs {
			// whether exist
			disk, err := GetDisksByLoc(loc)
			if err != nil {
				//TODO 	        AddLog(err)
			}

			err = disk.format()
			//TODO 	        AddLog(err)
		}
		return
	}
	res = "?????????"
	return
}

func (d *Disks) format() (err error) {
	o := orm.NewOrm()

	if d.Host != HOST_USED && d.Host != HOST_FOREIGN {
		err = fmt.Errorf("need not format")
		//TODO 	        AddLog(err)
	}

	// init old disk
	dd := make([]string, 0)
	block := make([]string, 0)

	dd = append(dd, "dd", "if=/dev/zero", "of=", "%s", "bs=4K", "count=16384", "oflag=direct")
	block = append(block, "blockdev", "--rereadpt", "")

	if res, err := util.Execute("/bin/sh", dd); err != nil {
		fmt.Println(res, "\n", err)
	}
	if res, err := util.Execute("/bin/sh", block); err != nil {
		fmt.Println(res, "\n", err)
		//TODO 	        AddLog(err)
		return err
	}

	// init new disk
	d.initNewDisk()

	// Delete disk by uuid
	if _, err = o.QueryTable(new(Disks)).Filter("uuid", d.Id).Delete(); err != nil {
		//TODO          AddLog(err)
		return
	}

	return
}

func removeDev() {
}

func (d *Disks) initNewDisk() (err error) {
	//log.info('init new disk %s(%s)...' % (self.location, self.dev_name))
	//speedio INFO init new disk 1.1.7(sde)...

	time.Sleep(1 * time.Second)
	//log.info('create rqr dm, cost:%s secs' % t.secs)
	//use golang timer to count some times TODO
	return
}

// Use location to get disk's info
func GetDisksByLoc(loc string) (d Disks, err error) {
	o := orm.NewOrm()

	if err = o.QueryTable(new(Disks)).Filter("location", loc).One(&d); err != nil {
		fmt.Printf("%+v\n\n!!!!!", err)
		if err.Error() == "<QuerySeter> no row found" {
			err = fmt.Errorf(" not exist")
			//TODO          AddLog(err)
			return
		}
		return
		//TODO          AddLog(err)
	}

	return
}

/*
var (
	Objects map[string]*Object
)

type Object struct {
	ObjectId   string
	Score      int64
	PlayerName string
}

func init() {
	Objects = make(map[string]*Object)
	Objects["hjkhsbnmn123"] = &Object{"hjkhsbnmn123", 100, "astaxie"}
	Objects["mjjkxsxsaa23"] = &Object{"mjjkxsxsaa23", 101, "someone"}
}

func AddOne(object Object) (ObjectId string) {
	object.ObjectId = "astaxie" + strconv.FormatInt(time.Now().UnixNano(), 10)
	Objects[object.ObjectId] = &object
	return object.ObjectId
}

func GetOne(ObjectId string) (object *Object, err error) {
	if v, ok := Objects[ObjectId]; ok {
		return v, nil
	}
	return nil, errors.New("ObjectId Not Exist")
}

func GetAll() map[string]*Object {
	return Objects
}

func Update(ObjectId string, Score int64) (err error) {
	if v, ok := Objects[ObjectId]; ok {
		v.Score = Score
		return nil
	}
	return errors.New("ObjectId Not Exist")
}

func Delete(ObjectId string) {
	delete(Objects, ObjectId)
}
*/
