package models

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/astaxie/beego/orm"
	"speedio/metadata"
	"speedio/models/lm"
	"speedio/util"
)

type DbDisk struct {
	Id        string    `orm:"column(uuid);size(255);pk"               json:"uuid"`
	Created   time.Time `orm:"column(created_at);type(datetime)"       json:"created_at"`
	Updated   time.Time `orm:"column(updated_at);type(datetime)"       json:"updated_at"`
	Loc       string    `orm:"column(location);size(255)"              json:"location"`
	Ploc      string    `orm:"column(prev_location);size(255)"         json:"prev_location"`
	Health    string    `orm:"column(health);size(255)"                json:"health"`
	Role      string    `orm:"column(role);size(255)"                  json:"role"`
	Raid      *DbRaid   `orm:"column(raid);size(255);rel(fk);null"     json:"raid"` //raid's uuid
	RaidHint  string    `orm:"column(raid_hint);size(255)"             json:"raid_hint"`
	Disktype  string    `orm:"column(disktype);size(255)"              json:"disktype"`
	Vendor    string    `orm:"column(vendor);size(255)"                json:"vendor"`
	Model     string    `orm:"column(model);size(255)"                 json:"model"`
	Host      string    `orm:"column(host);size(255)"                  json:"host"`
	Sn        string    `orm:"column(sn);size(255)"                    json:"sn"`
	CapSector int64     `orm:"column(cap_sector)"                      json:"cap_sector"`
	DevNo     int64     `orm:"column(devno)"                           json:"devno"`
	DevName   string    `orm:"column(dev_name);size(255)"              json:"dev_name"`
	UnplugSeq int64     `orm:"column(unplug_seq)"                      json:"unplug_seq"`
	Link      bool      `orm:"column(link)"                            json:"link"`
}

// api output
type ResDisk struct {
	Uuid      string  `json:"id"`
	Health    string  `json:"health"`
	Role      string  `json:"role"`
	Location  string  `json:"location"`
	Raid      string  `json:"raid"`
	CapSector int64   `json:"cap_sector"`
	CapMb     float64 `json:"cap_mb"`
	Vendor    string  `json:"vendor"`
	Model     string  `json:"model"`
	Sn        string  `json:"sn"`
	//rpm,rqr_count
}

type Disk struct {
	DbDisk
	//RaidId string `json:"raid_name"` //raid's name
}

type Uint64R []uint64

type ByUnplugSeq []DbDisk

func (u ByUnplugSeq) Len() int           { return len(u) }
func (u ByUnplugSeq) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u ByUnplugSeq) Less(i, j int) bool { return u[i].UnplugSeq > u[j].UnplugSeq }

func (a Uint64R) Len() int           { return len(a) }
func (a Uint64R) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Uint64R) Less(i, j int) bool { return a[i] < a[j] }

var (
	HOST_LOCAL   = "local" //db has no this disk's data, but metadata host uuid equal with storage
	HOST_FOREIGN = "foreign"
	HOST_USED    = "used"
)

func (d *DbDisk) TableName() string {
	return "disks"
}

func init() {
	orm.RegisterModel(new(DbDisk))

	// scan_all
	/*util.AddLog("system start scan")
	if util.CheckSystemRaid1() {
		//	make_system_rd()
		//	        check_system_rd()
	}*/

	//ScanAll()
}

func ScanAllDisks() (err error) {
	util.AddLog("disk scan...")
	disks, err := DiskAll()
	if err != nil {
		util.AddLog(err)

	}
	for _, disk := range disks {
		if disk.Host != HOST_NATIVE {
			disk.Delete()
		}
	}
	//TODO part

	var m lm.LocationMapping
	m.Init()
	count := len(m.Mapping)

	for {
		//	time.Sleep(16*time.Second)
		time.Sleep(0 * time.Second)
		m.Init()
		if count == len(m.Mapping) {
			break
		}
		count = len(m.Mapping)
		util.AddLog("new disk online, wait again")
	}

	// TODO ssd
	mapping := m.Mapping

	// Timer
	start := time.Now()
	//log.warning('mapping: %s' % mapping)
	util.AddLog(fmt.Sprintf("mapping: %s", m.Mapping))
	if len(mapping) == 0 {
		err = fmt.Errorf("no any disks")
		return
	}

	for loc, devName := range mapping {
		var disk DbDisk
		o := orm.NewOrm()
		if err = o.QueryTable(new(DbDisk)).Filter("dev_name", devName).Filter("location", loc).One(&disk); err != nil {
			disk.Loc = loc
			disk.DevName = devName
		}

		// TODO parts base
		disk.Scan() // TODO goroutine

	}
	elapsed := time.Since(start)

	cmd := "sync"
	util.ExecuteByStr(cmd, false)
	util.AddLog(fmt.Sprintf("scan all disks, cost:%s secs", elapsed))

	return
}

// Scan
func (d *DbDisk) Scan() (disk DbDisk) {
	d._grabDiskVpd()
	//log.warning('scan disk %s(%s)...' % (self.location, self.dev_name))
	host, md := d.classify()

	fmt.Println(host)
	switch host {
	case "native":
		d.initNativeDisk(md)
	case "local":
		d.initLocalDisk(md)
	case "foreign":
		d.initForeignDisk(md)
	case "used":
		d.initUsedDisk()
	case "new":
		d.initNewDisk()
	}

	d.Save(map[string]interface{}{"prev_location": d.Loc})
	return
}

func (d *DbDisk) initNewDisk() {
	util.AddLog(fmt.Sprintf("init new disk %s(%s)...", d.Loc, d.DevName))

	//parts := "" // TODO
	//dmcreate("s"+d.DevName, parts)

	start := time.Now()
	md := metadata.NewDiskMetadata_1()
	md.Attrs["disk_uuid"] = util.Uuid4()
	cond := map[string]string{"host_uuid": md.Attrs["host_uuid"], "disk_uuid": md.Attrs["disk_uuid"]}

	if _, err := metadata.Write("/dev/"+d.DevName, 1, cond); err != nil {
		util.AddLog(err)
	}

	elapsed := time.Since(start)
	util.AddLog(fmt.Sprintf("save %s md, cost:%s secs", d.DevName, elapsed))

	d.Id = md.Attrs["disk_uuid"]
	d.Host = HOST_NATIVE

	// dev's st_rdev
	stat := syscall.Stat_t{}
	if err := syscall.Stat(d.DevPath(), &stat); err != nil {
		util.AddLog(err)
	}
	d.DevNo = int64(stat.Rdev)

	util.AddLog("finish scan")
	// return parts?
}

func (d *DbDisk) initNativeDisk(md *metadata.DiskMetadata_1) {
	util.AddLog(fmt.Sprintf("init native disk %s(%s)...", d.Loc, d.DevName))
	d.Id = md.Attrs["disk_uuid"]
	d.Host = HOST_NATIVE
	// dev's st_rdev
	stat := syscall.Stat_t{}
	err := syscall.Stat(d.DevPath(), &stat)
	if err != nil {
		util.AddLog(err)
	}
	d.DevNo = int64(stat.Rdev)

	if d.Health != HEALTH_FAILED {
		d.Health = HEALTH_NORMAL
	}
	// TODO parts
}

func (d *DbDisk) initLocalDisk(md *metadata.DiskMetadata_1) {
	d.initForeignDisk(md)
}

func (d *DbDisk) initForeignDisk(md *metadata.DiskMetadata_1) {
	util.AddLog(fmt.Sprintf("init foreign disk %s(%s)...", d.Loc, d.DevName))

	//TODO parts

	// dev's st_rdev
	stat := syscall.Stat_t{}
	err := syscall.Stat(d.DevPath(), &stat)
	if err != nil {
		util.AddLog(err)
	}

	d.Id = md.Attrs["disk_uuid"]
	d.Host = HOST_FOREIGN
	d.DevNo = int64(stat.Rdev)
	d.RaidHint = md.Attrs["raid_uuid"]
	fmt.Printf("%+v", d)
}

func (d *DbDisk) initUsedDisk() {
	util.AddLog(fmt.Sprintf("init used disk %s(%s)...", d.Loc, d.DevName))

	d.Health = HEALTH_NORMAL
	d.Id = util.Uuid4()
	d.Host = HOST_USED

	// dev's st_rdev
	stat := syscall.Stat_t{}
	err := syscall.Stat(d.DevPath(), &stat)
	if err != nil {
		util.AddLog(err)
	}
	d.DevNo = int64(stat.Rdev)
	util.AddLog("finish scan")
}

// Scan
func (d *DbDisk) classify() (string, *metadata.DiskMetadata_1) {
	if len(d.partitions()) == 0 {
		md, err := metadata.Parse(d.MdPath(), 1)
		if err == nil {
			if md.Attrs["host_uuid"] == util.HostUuid() {
				fmt.Println("111")
				if d.Uuid() == md.Attrs["disk_uuid"] {
					return "native", md
				} else {
					return "local", md
				}
			} else {
				return "foreign", md
			}
		} else {
			return "new", nil
		}
	} else {
		return "used", nil
	}

}

// Scan
func (d *DbDisk) partitions() (devs []string) {
	nr, err := filepath.Glob(fmt.Sprintf("/dev/%s*", d.DevName))
	if err != nil {
		util.AddLog(err)
	}

	part := make(map[uint64]string, 0)
	for _, n := range nr {
		dev := filepath.Base(n)
		if d.DevName == dev {
			continue
		}
		// dev's st_rdev
		stat := syscall.Stat_t{}
		err := syscall.Stat("/dev/"+dev, &stat)
		if err != nil {
			util.AddLog(err)

		}
		part[stat.Rdev] = dev

	}

	var keys Uint64R
	for k, _ := range part {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	for _, k := range keys {
		devs = append(devs, part[k])
	}
	return devs
}

// Scan
func (d *DbDisk) _grabDiskVpd() (err error) {
	d.Disktype = DISKTYPE_SATA
	d.Vendor = "UNKOWN"
	d.Model = "UNKOWN"
	d.Sn = ""
	d.CapSector = 0 //Sector(0)

	cmd := fmt.Sprintf("hdparm -I /dev/%s", d.DevName)
	o, _ := util.ExecuteByStr(cmd, true)

	m := regexp.MustCompile(`Model Number:\s*(.+)`)
	if m.MatchString(o) {
		mStr := m.FindSubmatch([]byte(o))
		if len(mStr) == 2 {
			mList := strings.Fields(string(mStr[1]))
			if len(mList) == 2 {
				d.Vendor = mList[0]
				d.Model = mList[1]
			}
		}
	}

	mm := regexp.MustCompile(`Serial Number:\s*(.+)`)
	if mm.MatchString(o) {
		mmStr := mm.FindSubmatch([]byte(o))
		if len(mmStr) == 2 {

			d.Sn = strings.TrimSpace(string(mmStr[1]))
		}
	}

	cmd = fmt.Sprintf("blockdev --getsz %s", d.DevPath())
	sizes, _ := util.ExecuteByStr(cmd, true)
	size := strings.TrimSpace(sizes)

	if len(size) == 0 {
		d.CapSector = 0
	} else {
		d.CapSector, err = strconv.ParseInt(strings.TrimSpace(sizes), 10, 64)
		if err != nil {
			util.AddLog(err)
			return
		}
	}
	return
}

// GET
// Get all disks
// TODO need filter condition
func GetRestDisks() (res []ResDisk, err error) {
	o := orm.NewOrm()

	var ds []DbDisk
	res = make([]ResDisk, 0)
	if _, err = o.QueryTable(new(DbDisk)).All(&ds); err != nil { //TODO
		util.AddLog(err)
		return
	}

	for _, disk := range ds {

		var re ResDisk
		re.Uuid = disk.Id
		re.Health = disk.Health
		re.Role = disk.Role
		re.Location = disk.Loc
		re.CapSector = disk.CapSector
		re.CapMb = float64(disk.CapSector / 1024 / 1024 / 2)
		re.Vendor = disk.Vendor
		re.Model = disk.Model
		re.Sn = disk.Sn
		if disk.Raid == nil {
			re.Raid = ""
		} else {
			// Get Foreign Key
			if _, err = o.LoadRelated(&disk, "Raid"); err != nil {
				util.AddLog(err)
			}
			re.Raid = disk.Raid.Name
		}
		res = append(res, re)
	}

	return
}

// DELETE
// Delete disks
func (d *DbDisk) Delete() {
	o := orm.NewOrm()
	if num, err := o.Delete(d); err == nil {
		fmt.Println(num)
	}
}

// PUT
// Format disks
// Del used disks then create new one
func FormatDisks(locations string) (res string, err error) {
	o := orm.NewOrm()
	util.AddLog(fmt.Sprintf("==== formating disk, locations:%s ====", locations))

	// when loc == all
	// format all
	var d Disk
	if locations == "all" {
		var disks []DbDisk
		if _, err = o.QueryTable(new(DbDisk)).All(&disks); err != nil { //TODO
			util.AddLog(err)
			return "", err
		}

		for _, disk := range disks {
			if disk.Host == HOST_USED || disk.Host == HOST_FOREIGN {
				d.DbDisk = disk
				if err = d.Format(); err != nil {
					util.AddLog(err)
					return "", err
				}
			}
		}

		// format single disk
	} else {
		locs := strings.FieldsFunc(locations, func(c rune) bool { return c == ',' })
		for _, loc := range locs {
			// lookup
			item_disk := map[string]interface{}{"location": loc}
			disks, err := GetDisksByArgv(item_disk)
			if err != nil {
				util.AddLog(err)
				return "", err
			}
			for _, disk := range disks {
				d.DbDisk = disk
				if err = d.Format(); err != nil {
					util.AddLog(err)
					return "", err
				}
			}
		}
	}
	util.AddLog(fmt.Sprintf("==== complete formating disk, locations:%s ====", locations))
	return
}

// Disks format
func (d *Disk) Format() (err error) {
	o := orm.NewOrm()

	if d.Host != HOST_USED && d.Host != HOST_FOREIGN {
		err = fmt.Errorf("need not format")
		util.AddLog(err)
		return err
	}

	// remove dev
	d.RemoveDev()

	// init old disk
	dd := make([]string, 0)
	block := make([]string, 0)

	dd = append(dd, "dd", "if=/dev/zero", "of=", "%s", "bs=4K", "count=16384", "oflag=direct")
	block = append(block, "blockdev", "--rereadpt", "")

	if res, err := util.Execute("/bin/sh", dd, true); err != nil {
		fmt.Println(res, "\n", err)
	}
	if res, err := util.Execute("/bin/sh", block, true); err != nil {
		fmt.Println(res, "\n", err)
		//TODO 	        AddLog(err)
		return err
	}

	// init new disk
	d.initNewDisk()

	// Delete disk by uuid
	if _, err = o.QueryTable(new(DbDisk)).Filter("uuid", d.Id).Delete(); err != nil {
		//TODO          AddLog(err)
		return
	}
	return
}

// Disks format
func (d *Disk) RemoveDev() (err error) {
	dmPath := fmt.Sprintf("/dev/mapper/s%s", d.DevName)

	if _, err = os.Stat(dmPath); err == nil {
		if err = dmremove(dmPath); err != nil {
			return
		}
	}
	return
}

// Disks format
func dmremove(path string) (err error) {
	cmd := fmt.Sprintf("dmsetup remove %s", path)
	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	if err = ensure_not_exist(path); err != nil {
		return
	}
	return
}

// Disks format
func ensure_not_exist(path string) (err error) {
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 120; i++ {
		if _, err = os.Stat(path); os.IsNotExist(err) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if _, err = os.Stat(path); err == nil {
		util.AddLog("ensure " + path + " not exist error")
		return
	}
	return
}

// Disks format
func ensureExist(path string) (err error) {
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 120; i++ {
		if _, err = os.Stat(path); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if _, err = os.Stat(path); err == nil {
		//util.ReadFile(path)
		return
	} else {
		err = fmt.Errorf("ensure " + path + " exist error")
		util.AddLog(err)
	}
	return
}

func RemovePart(dev string) {

}

// lookup
// Get disks by any argv
func GetDisksByArgv(item map[string]interface{}) (ds []DbDisk, err error) {
	o := orm.NewOrm()
	//var cond *orm.Condition
	//	cond1 :=
	for k, v := range item {

		switch k {
		case "uuid":
			if exist := o.QueryTable(new(DbDisk)).Filter(k, v).Exist(); !exist {
				err = fmt.Errorf("not exist")
				util.AddLog(err)
				return
			}
		case "location":
			locs := strings.FieldsFunc(v.(string), func(c rune) bool { return c == ',' }) //Dangerous
			for _, loc := range locs {
				if exist := o.QueryTable(new(DbDisk)).Filter(k, loc).Exist(); !exist {
					err = fmt.Errorf("not exist")
					util.AddLog(err)
					return
				}
				var temp DbDisk
				if err = o.QueryTable(new(DbDisk)).Filter(k, loc).One(&temp); err != nil {
					util.AddLog(err)
					return
				}
				ds = append(ds, temp)
			}
		default:
			if exist := o.QueryTable(new(DbDisk)).Filter(k, v).Exist(); !exist {
				err = fmt.Errorf("not exist")
				util.AddLog(err)
				return
			}
			if _, err = o.QueryTable(new(DbDisk)).Filter(k, v).All(&ds); err != nil {
				util.AddLog(err)
				return
			}
		}
	}
	return
}

// UPDATE
// Save disk
// Update Disk's infos
func (d *DbDisk) Save(items map[string]interface{}) (err error) {
	o := orm.NewOrm()

	// TODO force
	force := false

	// checking value
	for k, v := range items {
		switch k {
		case "uuid":
			d.Id = v.(string)
		case "location":
			d.Loc = v.(string)
		case "prev_location":
			d.Ploc = v.(string)
		case "health":
			d.Health = v.(string)
		case "role":
			d.Role = v.(string)
		case "raid":
			if v == nil {
				var temp DbDisk
				d.Raid = temp.Raid
			} else {
				r, err := GetRaidsByArgv(map[string]interface{}{"uuid": v})
				if err != nil {
					util.AddLog(err)
					return err // TODO thinking !!!!!!!!!
				}
				d.Raid = &r.DbRaid
			}
		case "disktype":
			d.Disktype = v.(string)
		case "vendor":
			d.Vendor = v.(string)
		case "model":
			d.Model = v.(string)
		case "host":
			d.Host = v.(string)
		case "sn":
			d.Sn = v.(string)
		case "cap_sector":
			d.CapSector = int64(v.(int))
		case "dev_name":
			d.DevName = v.(string)
		case "unplug_seq":
			d.UnplugSeq = int64(v.(int))
		case "link":
			d.Link = v.(bool)
		}
	}

	// Default
	if len(d.Health) == 0 {
		d.Health = HEALTH_NORMAL
	}
	if len(d.Role) == 0 {
		d.Role = ROLE_UNUSED
	}

	if !d.Exist() || force {
		d.Created = time.Now()
		d.Updated = time.Now()

		if _, err = o.Insert(d); err != nil {
			util.AddLog(err)
			return
		}
	} else {
		// Keep Created
		if d.Created.IsZero() {
			d.Created = time.Now()
		}

		d.Updated = time.Now()
		if _, err = o.Update(d); err != nil {
			util.AddLog(err)
			return
		}
	}

	return
}

// Get Uuid
func (d *DbDisk) Uuid() string {
	o := orm.NewOrm()
	var disk DbDisk
	if err := o.QueryTable(new(DbDisk)).Filter("dev_name", d.DevName).Filter("location", d.Loc).One(&disk); err != nil {
		util.AddLog(err)
		return ""
	}

	if disk.Id != "" {
		return disk.Id
	}

	return disk.Id
}

// exist
func (d *DbDisk) Exist() bool {
	_, err := GetDisksByArgv(map[string]interface{}{"uuid": d.Id})
	if err != nil {
		return false
	}
	return true
}

// Get disk's online status
func (d *DbDisk) Online() bool {
	// path.exist
	f, err := os.Stat(d.DevPath())
	if os.IsNotExist(err) {
		return false
	}
	// is not dir
	return !f.IsDir()
}

// Get disk's dev_path
func (d *DbDisk) DevPath() (dm_path string) {
	dm_path = fmt.Sprintf("/dev/mapper/s%s", d.DevName)
	if _, err := os.Stat(dm_path); os.IsNotExist(err) {
		return fmt.Sprintf("/dev/%s", d.DevName)
	}
	return
}

// md_path
func (d *DbDisk) MdPath() string {
	return "/dev/" + d.DevName
}

func DiskAll() (ds []DbDisk, err error) {
	o := orm.NewOrm()
	if _, err = o.QueryTable(new(DbDisk)).All(&ds); err != nil {
		util.AddLog(err)
		return
	}
	return
}
