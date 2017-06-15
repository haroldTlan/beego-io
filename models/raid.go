package models

import (
	"fmt"
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
	DevPath    string
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

	cmd_1 := make([]string, 0)
	cmd_1 = append(cmd_1, "if=/dev/zero", "of="+r.odevPath(), "bs=1M", "count=128", "oflag=direct")
	if _, err = util.Execute("dd", cmd_1); err != nil {
		util.AddLog(err)
		return
	}
	util.AddLog(strings.Join(cmd_1, " "))

	cmd_2 := make([]string, 0)
	cmd_2 = append(cmd_2, r.odevPath(), "-ff", "-y", "--metadatacopies", "1")
	if _, err = util.Execute("pvcreate", cmd_2); err != nil {
		util.AddLog(err)
		return
	}
	util.AddLog(strings.Join(cmd_2, " "))

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
				util.AddLog(err)
				return
			}
			if err = o.QueryTable(new(Raids)).Filter(k, v).One(&r); err != nil {
				util.AddLog(err)
				return
			}

		}
	}
	// when items > 0 TODO

	return
}

// Update extents
func (r *Raid) updateExtents() (err error) {
	if r.health() == HEALTH_DEGRADED || r.health() == HEALTH_NORMAL {
		cmd := make([]string, 0)
		cmd = append(cmd, r.odevPath(), "-o", "pv_pe_alloc_count,pv_pe_count")
		output, _ := util.Execute("pvs", cmd)
		util.AddLog(strings.Join(cmd, " "))

		caps := strings.Fields(output)

		r.UsedCap, _ = strconv.Atoi(caps[len(caps)-2])
		r.Cap, _ = strconv.Atoi(caps[len(caps)-1])

	}
	return
}

// Join Vg
func (r *Raid) joinVg() (err error) {
	cmd := make([]string, 0)
	cmd = append(cmd, "-o", "vg_name")
	output, _ := util.Execute("vgs", cmd)
	util.AddLog(strings.Join(cmd, " "))

	vgsMap := make(map[string]bool, 0)
	for _, vgs := range strings.Fields(output) {
		vgsMap[vgs] = true
	}

	// when vgName in output
	if ok := vgsMap[r.vgName()]; ok {
		cmd := make([]string, 0)
		cmd = append(cmd, r.vgName(), r.odevPath())
		if _, err = util.Execute("vgextend", cmd); err != nil {
			return
		}
		util.AddLog(strings.Join(cmd, " "))
	} else {
		cmd := make([]string, 0)
		cmd = append(cmd, "-s", "1024m", r.vgName(), r.odevPath()) //1024m TODO
		if _, err = util.Execute("vgextend", cmd); err != nil {
			return
		}
		util.AddLog(strings.Join(cmd, " "))
	}

	return
}

// Mdadm created
// Raid's method
func (r *Raid) mdadmCreate() (err error) {
	u := strings.Join(strings.Split(r.Id, "-"), "")
	mdadmUuid := fmt.Sprintf("%s-%s-%s-%s", u[0:8], u[8:16], u[16:24], u[24:])

	var raid_disk_paths []string
	for _, d := range r.RaidDisks {
		raid_disk_paths = append(raid_disk_paths, d.DevPath())
	}

	var sync string
	cmd := make([]string, 0)

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
		cmd = append(cmd, "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", "--force", "-q", "--name="+r.Name)
	} else if r.Level == 1 || r.Level == 10 {
		cmd = append(cmd, "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", sync, "--force", "-q", "--name="+r.Name,
			"--bitmap=/home/zonion/bitmap/"+bitmap, "--bitmap-chunk=16M")
	} else {
		cmd = append(cmd, "--create", r.devPath(), "--homehost=speedio",
			"--uuid="+mdadmUuid, "--level="+level, "--chunk=256", "--raid-disks="+count,
			strings.Join(raid_disk_paths, " "), "--run", sync, "--force", "-q", "--name="+r.Name,
			"--bitmap=/home/zonion/bitmap/"+bitmap, "--bitmap-chunk=16M", "--layout=left-symmetric")
	}

	util.AddLog(strings.Join(cmd, " "))
	if _, err = util.Execute("mdadm", cmd); err != nil {
		util.AddLog(err)
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

// Get raid's health
func (r *Raid) health() string {
	cmd := make([]string, 0)
	cmd = append(cmd, "--detail", r.devPath())
	output, err := util.Execute("mdadm", cmd)
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

func (r *Raid) active_rebuild_priority() (err error) {
	//TODO min

	cmd := make([]string, 0)
	if len(r.RaidDisks) < 12 {
		cmd = append(cmd, "20480", ">", "/sys/block/"+r.DevName+"/md/stripe_cache_size")
		if _, err = util.Execute("echo", cmd); err != nil {
			return
		}
		util.AddLog(strings.Join(cmd, " "))
	}

	cmd = make([]string, 0)
	cmd = append(cmd, "0", ">", "/sys/block/"+r.DevName+"/md/preread_bypass_threshold")
	if _, err = util.Execute("echo", cmd); err != nil {
		return
	}
	util.AddLog(strings.Join(cmd, " "))
	return
}

// TODO now when partitions=0
func (r *Raid) _clean_existed_partition() (err error) {
	cmd := make([]string, 0)
	cmd = append(cmd, "--rereadpt", r.devPath())

	if _, err = util.Execute("blockdev", cmd); err != nil {
		util.AddLog(err)
		return
	}
	util.AddLog(strings.Join(cmd, " "))

	return
}

// Get vg name TODO
func (r *Raid) vgName() string {
	return "VG-" + r.Name
}

// Get raid's odev_path
func (r *Raid) odevPath() string {
	if false {
		//cache
	}
	if false {
		//sdd
	}

	return r.devPath()
}

// Get raid's dev path
func (r *Raid) devPath() string {
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
