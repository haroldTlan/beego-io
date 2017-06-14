package models

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"speedio/models/util"
)

type Raids struct {
	Id              string    `orm:"column(uuid);size(255);pk"          json:"uuid"`
	Created         time.Time `orm:"column(created_at);type(datetime)"  json:"created_at"`
	Updated         time.Time `orm:"column(updated_at);type(datetime)"  json:"updated_at"`
	Name            string    `orm:"column(name);size(255)"		        json:"name"`
	Level           int       `orm:"column(level)"					    json:"level"`
	Chunk           int       `orm:"column(chunk_kb)"					json:"chunk_kb"`
	Health          string    `orm:"column(health);size(255)"           json:"health"`
	Cap             int       `orm:"column(cap)"					    json:"cap"`
	UsedCap         int       `orm:"column(used_cap)"				    json:"used_cap"`
	RebuildProgress int       `orm:"column(rebuild_progress)"           json:"rebuild_progress"`
	Rebuilding      bool      `orm:"column(rebuilding)"				    json:"rebuilding"`
	DevName         string    `orm:"column(dev_name);size(255)"         json:"dev_name"`
	Deleted         bool      `orm:"column(deleted)"				    json:"deleted"`
}

type Raid struct {
	Raids
	Sync    bool
	Cache   bool
	DevPath string
}

func init() {
	orm.RegisterModel(new(Raids))
}

// GET
// Get all raids
func GetAllRaids() (rs []Raids, err error) {
	o := orm.NewOrm()
	rs = make([]Raids, 0)

	if _, err = o.QueryTable(new(Raids)).All(&rs); err != nil {
		util.AddLog(err)
		return
	}

	return

}

// POST
// Create Raid
func AddRaids(name, level, raid, spare, chunk, rebuildPriority string, sync, cache bool) (err error) {
	o := orm.NewOrm()

	//TODO _Add check argv

	uuid := util.Urandom()
	devName, err := _next_dev_name()
	if err != nil {
		util.AddLog(err)
		return
	}

	var r Raid
	r.Id = uuid
	r.Name = name
	r.Level, _ = strconv.Atoi(level)
	r.DevName = devName
	r.Created = time.Now()
	r.Updated = time.Now()
	r.Sync = sync
	r.Cache = cache

	raidDisks, err := GetDisksByArgv(map[string]interface{}{"location": raid})
	if err != nil {
		util.AddLog(err)
		return
	}

	spareDisks, err := GetDisksByArgv(map[string]interface{}{"location": spare})
	if err != nil {
		util.AddLog(err)
		return
	}

	if err = r.mdadmCreate(raidDisks, spareDisks); err != nil {
		util.AddLog(err)
		return
	}

	if _, err = o.Insert(&r); err != nil {
		util.AddLog(err)
		return
	}

	if err = UpdateDiskByRole(raid, ROLE_DATA, name, uuid, true); err != nil {
		util.AddLog(err)
		return
	}
	if err = UpdateDiskByRole(spare, ROLE_SPARE, name, uuid, true); err != nil {
		util.AddLog(err)
		return
	}

	//	log.journal_info('Raid %s is created successfully.' % self.name,\
	//	                         '成功建立阵列 %s' % self.name.encode('utf8'))
	return
}

// DELETE
// Delete Raid
func DelRaids(name string) (err error) {
	o := orm.NewOrm()

	item_raid := map[string]interface{}{"name": name}
	item_disk := map[string]interface{}{"raid": name}

	disks, err := GetDisksByArgv(item_disk)
	if err != nil {
		util.AddLog(err)
		return
	}

	for _, d := range disks {
		if err = UpdateDiskByRole(d.Loc, ROLE_UNUSED, "", "", false); err != nil {
			util.AddLog(err)
			return
		}
	}

	r, err := GetRaidsByArgv(item_raid)
	if err != nil {
		util.AddLog(err)
		return

	}
	if _, err = o.Delete(&r); err != nil {
		util.AddLog(err)
		return
	}

	return
}

// PUT
// Update Raid's status
func UpdateRaid() (err error) {
	return
}

func GetRaidsByArgv(item map[string]interface{}, items ...map[string]interface{}) (r Raids, err error) {
	o := orm.NewOrm()

	if len(items) == 0 {
		for k, v := range item {
			if exist := o.QueryTable(new(Raids)).Filter(k, v).Exist(); !exist {
				//TODO          AddLog(err)
				err = fmt.Errorf("not exist")
				return
			}
			if err := o.QueryTable(new(Raids)).Filter(k, v).One(&r); err != nil {
				//TODO          AddLog(err)
			}

		}
	}
	// when items > 0 TODO

	return
}

// Mdadm created
// Raid's method
func (r Raid) mdadmCreate(raidDisks, spareDisks []Disks) (err error) {
	u := strings.Join(strings.Split(r.Id, "-"), "")
	mdadmUuid := fmt.Sprintf("%s-%s-%s-%s", u[0:8], u[8:16], u[16:24], u[24:])

	var raid_disk_paths []string
	for _, d := range raidDisks {
		raid_disk_paths = append(raid_disk_paths, d.DevPath())
	}

	var sync string
	cmd := make([]string, 0)

	if r.Sync {
		sync = ""
	} else {
		sync = "--assume-clean"
	}

	bitmap := "@bitmap/" + r.Id + ".bitmap"

	//TODO	homehost := "speedio"
	//TODO chunk
	//TODO --bitmap-chunk
	//TODO --layout
	level := strconv.Itoa(r.Level)
	count := strconv.Itoa(len(raid_disk_paths))
	if r.Level == 0 {
		cmd = append(cmd, "mdadm", "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", "--force", "-q", "--name="+r.Name)
	} else if r.Level == 1 || r.Level == 10 {
		cmd = append(cmd, "mdadm", "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", sync, "--force", "-q", "--name="+r.Name,
			"--bitmap="+bitmap, "--bitmap-chunk=16M")
	} else {
		cmd = append(cmd, "mdadm", "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", sync, "--force", "-q", "--name="+r.Name,
			"--bitmap="+bitmap, "--bitmap-chunk=16M", "--layout=left-symmetric")
	}

	if _, err = util.Execute("/bin/sh", cmd); err != nil {
		util.AddLog(err)
		return
	}
	//TODO db.RaidRecovery.create

	//active_rebuild_priority(self.rebuild_priority)
	return
}

func (r Raid) devPath() string {
	return "/dev/" + r.DevName
}

// Get raid's dev name
func _next_dev_name() (string, error) {
	nr, err := filepath.Glob("/dev/md?*")
	if err != nil {
		return "md0", err
	}

	nrMap := make(map[string]bool, 0)
	for _, n := range nr {
		k := strings.Split(n, "md")[1]
		nrMap[k] = true
	}

	for i := 0; i < 100; i++ {
		num := strconv.Itoa(i)
		if ok := nrMap[num]; !ok {
			return fmt.Sprintf("md%s", num), nil
		}
	}

	return "md0", nil
}
