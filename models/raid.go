package models

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	Sync       bool
	Cache      bool
	RaidDisks  []Disks
	SpareDisks []Disks
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
	r.Chunk = 256

	r.RaidDisks, err = GetDisksByArgv(map[string]interface{}{"location": raid})
	if err != nil {
		util.AddLog(err)
		return
	}

	r.SpareDisks, err = GetDisksByArgv(map[string]interface{}{"location": spare})
	if err != nil {
		util.AddLog(err)
		return
	}

	// Mdadm
	if err = r.mdadmCreate(); err != nil {
		util.AddLog(err)
		return
	}

	//TODO create_ssd()
	//TODO create_cache()

	cmd_dd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=128 oflag=direct", r.OdevPath())
	if _, err = util.ExecuteByStr(cmd_dd); err != nil {
		return
	}

	cmd_pvcreate := fmt.Sprintf("pvcreate %s -ff -y --metadatacopies 1", r.OdevPath())
	if _, err = util.ExecuteByStr(cmd_pvcreate); err != nil {
		return
	}

	r.joinVg()
	r.updateExtents()

	if _, err = o.Insert(&r.Raids); err != nil {
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

	item_raid := map[string]interface{}{"name": name}
	r, err := GetRaidsByArgv(item_raid)
	if err != nil {
		util.AddLog(err)
		return
	}

	item_disk := map[string]interface{}{"raid": name}
	disks, err := GetDisksByArgv(item_disk)
	if err != nil {
		util.AddLog(err)
		return
	}

	if r.Online() {
		if r.Health() != HEALTH_FAILED {
			if err = r.detachVg(); err != nil {
				util.AddLog(err)
				return
			}

			cmd := fmt.Sprintf("pvremove %s", r.OdevPath())
			if _, err = util.ExecuteByStr(cmd); err != nil {
				return
			}

		}
		cmd := fmt.Sprintf("mdadm --stop %s", r.DevPath())
		if _, err = util.ExecuteByStr(cmd); err != nil {
			return
		}
	}

	for _, disk := range disks {
		if disk.Online() {
			cmd := fmt.Sprintf("mdadm --zero-superblock %s", disk.DevPath())
			if _, err = util.ExecuteByStr(cmd); err != nil {
				return
			}
		}
	}

	devDir := fmt.Sprintf("/dev/%s", r.VgName())
	cmd := fmt.Sprintf("rm %s -rf", devDir)
	if _, err = util.ExecuteByStr(cmd); err != nil {
		return
	}

	err = _DelRaids(name)
	if err != nil {
		return
	}
	return
}

// Delete raid from sqlite
func _DelRaids(name string) (err error) {
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
	if _, err = o.Delete(&r.Raids); err != nil {
		util.AddLog(err)
		return
	}
	return
}

// detachVg
func (r *Raid) detachVg() (err error) {
	cmd := "vgs -o pv_count"
	output, err := util.ExecuteByStr(cmd)
	if err != nil {
		return
	}

	var _cmd string
	nr, _ := strconv.Atoi(strings.Fields(output)[1])
	if nr == 1 {
		_cmd = fmt.Sprintf("vgremove %s", r.VgName())
	} else {
		_cmd = fmt.Sprintf("vgremove %s %s", r.VgName(), r.OdevPath())
	}
	if _, err = util.ExecuteByStr(_cmd); err != nil {
		return
	}

	return
}

// PUT
// Update Raid's status
func UpdateRaid() (err error) {
	return
}

func GetRaidsByArgv(item map[string]interface{}, items ...map[string]interface{}) (r Raid, err error) {
	o := orm.NewOrm()

	if len(items) == 0 {
		for k, v := range item {
			if exist := o.QueryTable(new(Raids)).Filter(k, v).Exist(); !exist {
				err = fmt.Errorf("not exist")
				util.AddLog(err)
				return
			}
			var raid Raids
			if err = o.QueryTable(new(Raids)).Filter(k, v).One(&raid); err != nil {
				util.AddLog(err)
				return
			}
			r.Raids = raid
		}
	}
	// when items > 0 TODO

	return
}

// func AddRaids
// Update extents
func (r *Raid) updateExtents() (err error) {
	if r.Health() == HEALTH_DEGRADED || r.Health() == HEALTH_NORMAL {
		cmd := fmt.Sprintf("pvs %s -o pv_pe_alloc_count,pv_pe_count", r.OdevPath())
		output, _ := util.ExecuteByStr(cmd)

		caps := strings.Fields(output)

		r.UsedCap, _ = strconv.Atoi(caps[len(caps)-2])
		r.Cap, _ = strconv.Atoi(caps[len(caps)-1])

	}
	return
}

// func AddRaids
// Join Vg
func (r *Raid) joinVg() (err error) {
	cmd := "vgs -o vg_name"
	output, _ := util.ExecuteByStr(cmd)

	vgsMap := make(map[string]bool, 0)
	for _, vgs := range strings.Fields(output) {
		vgsMap[vgs] = true
	}

	// when vgName in output
	if ok := vgsMap[r.VgName()]; ok {
		cmd := fmt.Sprintf("vgextend %s %s", r.VgName(), r.OdevPath())
		if _, err = util.ExecuteByStr(cmd); err != nil {
			return
		}
	} else {
		cmd := fmt.Sprintf("vgcreate -s 1024m %s %s", r.VgName(), r.OdevPath())
		if _, err = util.ExecuteByStr(cmd); err != nil {
			return
		}
	}

	return
}

// func AddRaids
// Mdadm created
func (r *Raid) mdadmCreate() (err error) {
	u := strings.Join(strings.Split(r.Id, "-"), "")
	mdadmUuid := fmt.Sprintf("%s-%s-%s-%s", u[0:8], u[8:16], u[16:24], u[24:])

	var raid_disk_paths []string
	for _, d := range r.RaidDisks {
		raid_disk_paths = append(raid_disk_paths, d.DevPath())
	}

	var sync, cmd string

	if r.Sync {
		sync = ""
	} else {
		sync = "--assume-clean"
	}

	bitmap := r.Id + ".bitmap"

	//TODO	homehost := "speedio"
	//TODO chunk
	//TODO --bitmap-chunk
	//TODO --layout
	level := strconv.Itoa(r.Level)
	count := strconv.Itoa(len(raid_disk_paths))
	if r.Level == 0 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run --force -q --name=\"%s\"",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), r.Name)
	} else if r.Level == 1 || r.Level == 10 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=/home/zonion/bitmap/%s --bitmap-chunk=16M",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)

	} else {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=/home/zonion/bitmap/%s --bitmap-chunk=16M --layout=left-symmetric",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)
	}

	if _, err = util.ExecuteByStr(cmd); err != nil {
		return
	}

	if err = r._clean_existed_partition(); err != nil {
		util.AddLog(err)
		return
	}
	//TODO db.RaidRecovery.create
	//partitions>0
	if r.Level == 5 || r.Level == 6 {
		r.active_rebuild_priority()
	}
	return
}

// Get raid's online status
func (r *Raid) Online() bool {
	f, err := os.Stat(r.DevPath())
	if os.IsNotExist(err) {
		return false // do not exist
	}
	// is not dir
	return !f.IsDir()
}

// Get raid's health
func (r *Raid) Health() string {
	cmd := fmt.Sprintf("mdadm --detail %s", r.DevPath())
	output, err := util.ExecuteByStr(cmd)
	if err != nil {
		return HEALTH_FAILED
	}

	re := regexp.MustCompile("State :\\s+([^\n]+)")
	segs := re.FindAllString(output, -1)

	status := strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' })
	sMap := make(map[string]bool, 0)
	for _, s := range status {
		sMap[s] = true
	}

	if ok := sMap["degraded"]; ok {
		return HEALTH_DEGRADED
	} else if ok := sMap["FAILED"]; ok {
		return HEALTH_FAILED
	} else {
		return HEALTH_NORMAL
	}

	return HEALTH_FAILED
}

// func mdadmCreate
func (r *Raid) active_rebuild_priority() (err error) {
	//TODO min

	if len(r.RaidDisks) < 12 {
		cmd := fmt.Sprintf("echo 20480 > /sys/block/%s/md/stripe_cache_size", r.DevName)
		if _, err = util.ExecuteByStr(cmd); err != nil {
			return
		}
	}
	cmd := fmt.Sprintf("echo 0 > /sys/block/%s/md/preread_bypass_threshold", r.DevName)
	if _, err = util.ExecuteByStr(cmd); err != nil {
		return
	}

	return
}

// func mdadmCreate
// TODO now when partitions=0
func (r *Raid) _clean_existed_partition() (err error) {
	cmd := fmt.Sprintf("blockdev --rereadpt %s", r.DevPath())
	if _, err = util.ExecuteByStr(cmd); err != nil {
		return
	}

	return
}

// Get vg name TODO
func (r *Raid) VgName() string {
	return "VG-" + r.Name
}

// Get raid's odev_path
func (r *Raid) OdevPath() string {
	if false {
		//cache
	}
	if false {
		//sdd
	}

	return r.DevPath()
}

// Get raid's dev path
func (r *Raid) DevPath() string {
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
