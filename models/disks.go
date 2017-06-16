package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"speedio/models/util"
)

type Disks struct {
	Id        string    `orm:"column(uuid);size(255);pk"          json:"uuid"`
	Created   time.Time `orm:"column(created_at);type(datetime)"  json:"created_at"`
	Updated   time.Time `orm:"column(updated_at);type(datetime)"  json:"updated_at"`
	Loc       string    `orm:"column(location);size(255)"         json:"location"`
	Ploc      string    `orm:"column(prev_location);size(255)"    json:"prev_location"`
	Health    string    `orm:"column(health);size(255)"           json:"health"`
	Role      string    `orm:"column(role);size(255)"             json:"role"`
	Raid      string    `orm:"column(raid);size(255)"             json:"raid"`    //raid's name
	RaidId    string    `orm:"column(raid_id);size(255)"          json:"raid_id"` //raid's uuid
	Disktype  string    `orm:"column(disktype);size(255)"         json:"disktype"`
	Vendor    string    `orm:"column(vendor);size(255)"           json:"vendor"`
	Model     string    `orm:"column(model);size(255)"            json:"model"`
	Host      string    `orm:"column(host);size(255)"             json:"host"`
	Sn        string    `orm:"column(sn);size(255)"               json:"sn"`
	CapSector int64     `orm:"column(cap_sector)"                 json:"cap_sector"`
	DevName   string    `orm:"column(dev_name);size(255)"         json:"dev_name"`
	Link      bool      `orm:"column(link)"                       json:"link"`
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
// Del used disks then create new one
func FormatDisks(locations string) (res string, err error) {
	util.AddLog(fmt.Sprintf("==== formating disk, locations:%s ====", locations))

	// when loc == all
	// format all
	if locations == "all" {
		disks, err := GetAllDisks()
		if err != nil {
			util.AddLog(err)
			return "", err
		}
		for _, disk := range disks {
			if disk.Host == HOST_USED || disk.Host == HOST_FOREIGN {
				if err = disk.Format(); err != nil {
					util.AddLog(err)
					return "", err
				}
			}
		}

		// format single disk
	} else {
		locs := strings.FieldsFunc(locations, func(c rune) bool { return c == ',' })
		for _, loc := range locs {
			item_disk := map[string]interface{}{"location": loc}
			disks, err := GetDisksByArgv(item_disk)
			if err != nil {
				util.AddLog(err)
				return "", err
			}
			for _, disk := range disks {
				err = disk.Format()
				if err != nil {
					util.AddLog(err)
					return "", err
				}
			}
		}
	}
	return
}

func (d *Disks) Format() (err error) {
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

// Get disks by any argv
func GetDisksByArgv(item map[string]interface{}, items ...map[string]interface{}) (d []Disks, err error) {
	o := orm.NewOrm()

	if len(items) == 0 {
		for k, v := range item {

			switch k {

			case "location":
				locs := strings.FieldsFunc(v.(string), func(c rune) bool { return c == ',' }) //Dangerous
				for _, loc := range locs {
					if exist := o.QueryTable(new(Disks)).Filter(k, loc).Exist(); !exist {
						//TODO          AddLog(err)
						err = fmt.Errorf("not exist")
						util.AddLog(err)
						return

					}
					var temp Disks
					if err = o.QueryTable(new(Disks)).Filter(k, loc).One(&temp); err != nil {
						util.AddLog(err)
						return
					}
					d = append(d, temp)
				}
			default:
				if exist := o.QueryTable(new(Disks)).Filter(k, v).Exist(); !exist {
					//TODO          AddLog(err)
					err = fmt.Errorf("not exist")
					util.AddLog(err)
					return
				}
				if _, err := o.QueryTable(new(Disks)).Filter(k, v).All(&d); err != nil {
					//TODO          AddLog(err)
					util.AddLog(err)
				}
			}
		}
	}
	// when items > 0 TODO

	return
}

// Update disk
// When raid's status has changed
func UpdateDiskByRole(locations, style, name, uuid string, link bool) (err error) {
	o := orm.NewOrm()

	loc := strings.Split(locations, ",")
	for _, l := range loc {
		d, err := GetDisksByLoc(l)
		if err != nil {
			//TODO          AddLog(err)
		}
		d.Role = style
		d.Raid = name
		d.RaidId = uuid
		d.Updated = time.Now()
		d.Link = link
		if _, err := o.Update(&d); err != nil {
			//TODO          AddLog(err)
		}
	}
	return
}

// Get disk's online status
func (d *Disks) Online() bool {
	f, err := os.Stat(d.DevPath())
	if os.IsNotExist(err) {
		return false // do not exist
	}
	// is not dir
	return !f.IsDir()
}

func (d *Disks) DevPath() (dm_path string) {
	dm_path = fmt.Sprintf("/dev/mapper/s%s", d.DevName)
	if _, err := os.Stat(dm_path); os.IsNotExist(err) {
		return fmt.Sprintf("/dev/%s", d.DevName)
	}
	return
}
