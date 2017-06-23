package models

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"speedio/models/util"
)

type DbRaid struct {
	Id         string    `orm:"column(uuid);size(255);pk"           json:"uuid"`
	Created    time.Time `orm:"column(created_at);type(datetime)"   json:"created_at"`
	Updated    time.Time `orm:"column(updated_at);type(datetime)"   json:"updated_at"`
	Name       string    `orm:"column(name);size(255)"		        json:"name"`
	Level      int64     `orm:"column(level)"					    json:"level"`
	Chunk      int64     `orm:"column(chunk_kb)"					json:"chunk_kb"`
	RbPriority string    `orm:"column(rebuild_priority);size(255)"  json:"rebuild_priority"`
	Health     string    `orm:"column(health);size(255)"            json:"health"`
	RdsNr      int64     `orm:"column(raid_disks_nr)"		        json:"raid_disks_nr"`
	SdsNr      int64     `orm:"column(spare_disks_nr)"		        json:"spare_disks_nr"`
	Cap        int64     `orm:"column(cap)"					        json:"cap"`
	UsedCap    int64     `orm:"column(used_cap)"				    json:"used_cap"`
	DevName    string    `orm:"column(dev_name);size(255)"          json:"dev_name"`
	OdevName   string    `orm:"column(odev_name);size(255)"         json:"odev_name"`
	UnplugSeq  int64     `orm:"column(unplug_seq)"                  json:"unplug_seq"`
	Deleted    bool      `orm:"column(deleted)"				        json:"deleted"`
	Disks      []*DbDisk `orm:"column(Id);reverse(many)"			json:"disks"`
}

func (r *DbRaid) TableName() string {
	return "raids"
}

type Raid struct {
	DbRaid
	Sync       bool
	Cache      bool
	RaidDisks  []Disk
	SpareDisks []Disk //resDisk
}

// output
type ResRaid struct {
	Rb         bool    `json:"rebuilding"`
	Uuid       string  `json:"id"`
	Health     string  `json:"health"`
	Level      int64   `json:"level"`
	Name       string  `json:"name"`
	Cap        int64   `json:"cap_sector"`
	Used       int64   `json:"used_cap_sector"`
	CapMb      float64 `json:"cap_mb"`
	UsedMb     float64 `json:"used_cap_mb"`
	ChunkKb    int     `json:"chunk_kb"`
	Blkdev     string  `json:"blkdev"`
	RbProgress float64 `json:"rebuild_progress"`
	//rqr_count
}

func init() {
	orm.RegisterModel(new(DbRaid))
}

// GET
// Get all raids
// TODO more condition to filter
func GetAllRaids() (res []ResRaid, err error) {
	o := orm.NewOrm()
	res = make([]ResRaid, 0)
	var rs []DbRaid

	if _, err = o.QueryTable(new(DbRaid)).Filter("deleted", false).All(&rs); err != nil {
		util.AddLog(err)
		return
	}

	for _, raid := range rs {
		var re ResRaid
		re.Uuid = raid.Id
		re.Level = raid.Level
		re.Name = raid.Name
		re.Cap = raid.Cap
		re.Used = raid.UsedCap
		re.Rb = false
		re.RbProgress = 0
		re.ChunkKb = 32
		re.Health = "normal"
		res = append(res, re)
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
	r.Level, _ = strconv.ParseInt(level, 10, 64)
	r.DevName = devName
	r.Created = time.Now()
	r.Updated = time.Now()
	r.Sync = sync
	r.Cache = cache
	r.Chunk = 256

	// init RaidDisks
	dataDisks, err := GetDisksByArgv(map[string]interface{}{"location": raid})
	if err != nil {
		util.AddLog(err)
		return
	}
	r.RdsNr = int64(len(dataDisks))
	/*dataDisks, err := r.RaidDisks(raid)
	if err != nil {
		util.AddLog(err)
		return
	}*/

	for _, disk := range dataDisks {
		var d Disk
		d.DbDisk = disk
		r.RaidDisks = append(r.RaidDisks, d)
	}

	// init SpareDisks
	spareDisks, err := GetDisksByArgv(map[string]interface{}{"location": spare})
	if err != nil {
		util.AddLog(err)
		return
	}
	r.SdsNr = int64(len(spareDisks))

	for _, disk := range spareDisks {
		var d Disk
		d.DbDisk = disk
		r.SpareDisks = append(r.SpareDisks, d)
	}

	// Mdadm
	if err = r.mdadmCreate(); err != nil {
		util.AddLog(err)
		return
	}

	//TODO create_ssd()
	//TODO create_cache()

	cmd_dd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=128", r.OdevPath())
	if _, err = util.ExecuteByStr(cmd_dd, true); err != nil {
		return
	}

	cmd_pvcreate := fmt.Sprintf("pvcreate %s -ff -y --metadatacopies 1", r.OdevPath())
	if _, err = util.ExecuteByStr(cmd_pvcreate, true); err != nil {
		return
	}

	r.joinVg()
	r.updateExtents()

	if _, err = o.Insert(&r.DbRaid); err != nil {
		util.AddLog(err)
		return
	}

	for _, disk := range dataDisks {
		if err = disk.Save(map[string]interface{}{"role": ROLE_DATA, "raid": uuid, "link": true}); err != nil {
			util.AddLog(err)
			return
		}
	}

	for _, disk := range spareDisks {
		if err = disk.Save(map[string]interface{}{"role": ROLE_DATA_SPARE, "raid": uuid, "link": true}); err != nil {
			util.AddLog(err)
			return
		}
	}

	util.AddLog(fmt.Sprintf("Raid %s is created successfully.", name))
	//	log.journal_info('Raid %s is created successfully.' % self.name,\
	//	                         '成功建立阵列 %s' % self.name.encode('utf8'))
	return
}

// DELETE
// Delete Raid
func DelRaids(name string) (err error) {
	o := orm.NewOrm()

	item_raid := map[string]interface{}{"name": name}
	r, err := GetRaidsByArgv(item_raid)
	if err != nil {
		util.AddLog(err)
		return
	}

	// get foreign keys
	if _, err = o.LoadRelated(&r.DbRaid, "Disks"); err != nil {
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
			if _, err = util.ExecuteByStr(cmd, true); err != nil {
				return
			}
		}

		cmd := fmt.Sprintf("mdadm --stop %s", r.DevPath())
		if _, err = util.ExecuteByStr(cmd, true); err != nil {
			return
		}
	}

	var d Disk
	for _, disk := range r.Disks {
		d.DbDisk = *disk
		if d.Online() {
			cmd := fmt.Sprintf("mdadm --zero-superblock %s", d.DevPath())
			if _, err = util.ExecuteByStr(cmd, true); err != nil {
				return
			}
		}
	}

	devDir := fmt.Sprintf("/dev/%s", r.VgName())
	cmd := fmt.Sprintf("rm %s -rf", devDir)
	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	err = _DelRaids(name)
	if err != nil {
		util.AddLog(err)
		return
	}
	return
}

// update raid from sqlite(deleted=true)
func _DelRaids(name string) (err error) {
	o := orm.NewOrm()

	item_raid := map[string]interface{}{"name": name}
	r, err := GetRaidsByArgv(item_raid)
	if err != nil {
		util.AddLog(err)
		return
	}

	// get foreign keys
	if _, err = o.LoadRelated(&r.DbRaid, "Disks"); err != nil {
		util.AddLog(err)
		return
	}

	for _, disk := range r.Disks {
		if err = disk.Save(map[string]interface{}{"role": ROLE_UNUSED, "raid": nil, "link": false}); err != nil {
			util.AddLog(err)
			return
		}
	}

	r.DbRaid.Save(map[string]interface{}{"deleted": true})
	/*r.DbRaid.Deleted = true
	if _, err = o.Update(&r.DbRaid); err != nil {
		util.AddLog(err)
		return
	}*/
	return
}

// DelRaids
func (r *DbRaid) detachVg() (err error) {
	cmd := "vgs -o pv_count"
	output, err := util.ExecuteByStr(cmd, true)
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
	if _, err = util.ExecuteByStr(_cmd, true); err != nil {
		return
	}

	return
}

// UPDATE
// Save raid
// Update Raid's infos
func (r *DbRaid) Save(item map[string]interface{}, items ...map[string]interface{}) (err error) {
	o := orm.NewOrm()

	if len(items) == 0 {
		//TODO k,v checking
		r._Save(item)
		r.Updated = time.Now()
		if _, err = o.Update(r); err != nil {
			util.AddLog(err)
			return
		}

	} else if len(items) > 0 {
		r._Save(item)
		for _, i := range items {
			r._Save(i)
		}
		r.Updated = time.Now()
		if _, err = o.Update(r); err != nil {
			util.AddLog(err)
			return
		}
	}

	return
}

func (r *DbRaid) _Save(item map[string]interface{}) {
	for k, v := range item {
		switch k {
		case "health":
			r.Health = v.(string)
		case "used_cap":
			r.UsedCap = v.(int64)
		case "odev_name":
			r.OdevName = v.(string)
		case "deleted":
			r.Deleted = v.(bool)
		}
	}
}

// Get raid disks
func (r *DbRaid) RaidDisks() (disks []DbDisk) {
	o := orm.NewOrm()

	// get foreign keys
	if _, err := o.LoadRelated(&r, "Disks"); err != nil {
		util.AddLog(err)
		return
	}
	for _, d := range r.Disks {
		if d.Role == ROLE_DATA {
			disks = append(disks, *d)
		}
	}

	return
}

// Get spare disks
func (r *DbRaid) SpareDisks() (disks []DbDisk) {
	o := orm.NewOrm()

	// get foreign keys
	if _, err := o.LoadRelated(&r, "Disks"); err != nil {
		util.AddLog(err)
		return
	}
	for _, d := range r.Disks {
		if d.Role == ROLE_DATA_SPARE {
			disks = append(disks, *d)
		}
	}

	return
}

// Get all disks in raid
func (r *DbRaid) AllDisks() (disks []*DbDisk) {
	o := orm.NewOrm()

	// get foreign keys
	if _, err := o.LoadRelated(&r, "Disks"); err != nil {
		util.AddLog(err)
		return
	}
	return r.Disks

}

// Has bitmap
func (r *DbRaid) hasBitmap() bool {
	filename := fmt.Sprintf("/sys/block/%s/md/bitmap/location", r.DevName)
	f := util.ReadFile(filename)

	return strings.TrimSpace(f) == "file"
}

// Create bitmap
func (r *DbRaid) createBitmap() {
	if r.supportBitmap() {
		bitmapFile := beego.AppConfig.String("bitmapfile")
		bitmapSize, _ := beego.AppConfig.Int("bitmapsize")

		bitmap := bitmapFile + r.Id + ".bitmap"
		cmd := fmt.Sprintf("mdadm --grow --bitmap=%s --bitmap-chunk=%sM %s", bitmap, bitmapSize, r.DevPath())
		util.ExecuteByStr(cmd, true)
	}
}

// Delete bitmap
func (r *DbRaid) deleteBitmap() {
	bitmapFile := beego.AppConfig.String("bitmapfile")
	bitmap1 := bitmapFile + r.Id + ".bitmap"
	bitmap2 := bitmapFile + ".disk/" + r.Id + ".bitmap"

	if _, err := os.Stat(bitmap1); err == nil {
		os.Remove(bitmap1)
	}
	if _, err := os.Stat(bitmap2); err == nil {
		os.Remove(bitmap2)
	}
}

// Support bitmap
func (r *DbRaid) supportBitmap() bool {
	lMap := make(map[int64]bool, 0)
	for _, i := range []int64{1, 5, 10} {
		lMap[i] = true
	}
	if ok := lMap[r.Level]; ok {
		return true
	}
	return false
}

// Delete cache
func (r *DbRaid) deleteCache() {
	cacheEnabled, _ := beego.AppConfig.Bool("cache_enable")
	if cacheEnabled {
		dmremove(r.OdevPath())
	}
}

// Create cache
func (r *DbRaid) createCache() {
	cacheEnabled, _ := beego.AppConfig.Bool("cacheenable")
	raidcachememsize, _ := beego.AppConfig.Int("raid_cache_mem_size")
	if !cacheEnabled {
		cmd := fmt.Sprintf("blockdev --getsz %s", r.OdevPath())
		size, err := util.ExecuteByStr(cmd, true)
		if err != nil {
			return
		}
		odev_name := "c" + r.DevName
		rule := fmt.Sprintf("0 %s cache %s %s %s %s", strings.TrimSpace(size), r.OdevPath(), raidcachememsize, r.Chunk, r.RdsNr) //TODO

		dmcreate(odev_name, rule)
		r.Save(map[string]interface{}{"odev_name": odev_name})
	}
}

// speedio lookup
// return Raid
func GetRaidsByArgv(item map[string]interface{}, items ...map[string]interface{}) (r Raid, err error) {
	o := orm.NewOrm()

	if len(items) == 0 {
		for k, v := range item {
			if exist := o.QueryTable(new(DbRaid)).Filter(k, v).Exist(); !exist {
				err = fmt.Errorf("not exist")
				util.AddLog(err)
				return
			}

			var raid DbRaid
			if err = o.QueryTable(new(DbRaid)).Filter(k, v).Filter("deleted", false).One(&raid); err != nil {
				util.AddLog(err)
				return
			}
			r.DbRaid = raid
		}
	}

	return
}

// func AddRaids
// Update extents
func (r *Raid) updateExtents() (err error) {
	if r.Health() == HEALTH_DEGRADED || r.Health() == HEALTH_NORMAL {
		cmd := fmt.Sprintf("pvs %s -o pv_pe_alloc_count,pv_pe_count", r.OdevPath())
		output, _ := util.ExecuteByStr(cmd, true)

		caps := strings.Fields(output)

		r.UsedCap, _ = strconv.ParseInt(caps[len(caps)-2], 10, 64)
		r.Cap, _ = strconv.ParseInt(caps[len(caps)-1], 10, 64)

	}
	return
}

// func AddRaids
// Join Vg
func (r *DbRaid) joinVg() (err error) {
	cmd := "vgs -o vg_name"
	output, _ := util.ExecuteByStr(cmd, true)

	vgsMap := make(map[string]bool, 0)
	for _, vgs := range strings.Fields(output) {
		vgsMap[vgs] = true
	}

	// when vgName in output
	if ok := vgsMap[r.VgName()]; ok {
		cmd := fmt.Sprintf("vgextend %s %s", r.VgName(), r.OdevPath())
		if _, err = util.ExecuteByStr(cmd, true); err != nil {
			return
		}
	} else {
		cmd := fmt.Sprintf("vgcreate -s 1024m %s %s", r.VgName(), r.OdevPath())
		if _, err = util.ExecuteByStr(cmd, true); err != nil {
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

	bitmapFile := beego.AppConfig.String("bitmapfile")
	//bitmapFile := "/home/zonion/bitmap/"
	bitmap := bitmapFile + r.Id + ".bitmap"

	//TODO	homehost := "speedio"
	//TODO chunk
	//TODO --bitmap-chunk
	//TODO --layout
	level := strconv.FormatInt(r.Level, 10)
	count := strconv.Itoa(len(raid_disk_paths))
	if r.Level == 0 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run --force -q --name=\"%s\"",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), r.Name)
	} else if r.Level == 1 || r.Level == 10 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=%s --bitmap-chunk=16M",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)

	} else {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=%s --bitmap-chunk=16M --layout=left-symmetric",
			r.DevPath(), mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)
	}

	if _, err = util.ExecuteByStr(cmd, true); err != nil {
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
// Whether exist and not a directory
func (r *DbRaid) Online() bool {
	f, err := os.Stat(r.DevPath())
	if os.IsNotExist(err) {
		return false // do not exist
	}
	// is not dir
	return !f.IsDir()
}

// Get raid's health
func (r *Raid) Health() string {
	time.Sleep(1 * time.Second)
	cmd := fmt.Sprintf("mdadm --detail %s", r.DevPath())
	output, err := util.ExecuteByStr(cmd, true)
	if err != nil {
		return HEALTH_FAILED
	}

	re := regexp.MustCompile(`State :\s+([^\n]+)`)
	segs := re.FindAllString(output, -1)

	// split by ":"
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

// func mdadmCreate TODO change method
func (r *Raid) active_rebuild_priority() (err error) {
	//TODO min

	if len(r.RaidDisks) < 12 {
		cmd := fmt.Sprintf("echo 20480 > /sys/block/%s/md/stripe_cache_size", r.DevName)
		if _, err = util.ExecuteByStr(cmd, true); err != nil {
			return
		}
	}
	cmd := fmt.Sprintf("echo 0 > /sys/block/%s/md/preread_bypass_threshold", r.DevName)
	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	return
}

// func mdadmCreate
// TODO now when partitions=0
func (r *DbRaid) _clean_existed_partition() (err error) {
	cmd := fmt.Sprintf("blockdev --rereadpt %s", r.DevPath())
	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	return
}

// Get vg name TODO
func (r *DbRaid) VgName() string {
	return "VG-" + r.Name
}

// Get raid's odev_path
func (r *DbRaid) OdevPath() string {
	if false {
		//cache
	}
	if false {
		//sdd
	}

	return r.DevPath()
}

// Get raid's dev path
func (r *DbRaid) DevPath() string {
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

func dmcreate(name string, rules string) {
	tmpfile := "/tmp/" + name + ".rule"

	//Write(tmpfile, strings.Join(rules, "\n"))

	cmd := fmt.Sprintf("dmsetup create %s %s", name, tmpfile)
	util.ExecuteByStr(cmd, true)

	dm_path := "/dev/mapper/" + name
	ensureExist(dm_path)

	os.Remove(tmpfile)
}
