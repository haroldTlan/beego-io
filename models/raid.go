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

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"speedio/util"
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
func AddRaids(name, raid, spare, rebuildPriority string, chunk, level int64, sync bool) (err error) {
	o := orm.NewOrm()

	err = CheckRaidsByArgv(map[string]interface{}{"name": name, "level": level, "raid_disks": raid, "spare_disks": spare, "chunk_kb": chunk, "rebuild_priority": rebuildPriority})
	if err != nil {
		util.AddLog(err)
		return
	}

	uuid := util.Urandom()
	devName, err := nextDevName()
	if err != nil {
		util.AddLog(err)
		return
	}

	var r Raid
	r.Id = uuid
	r.Name = name
	r.Level = level
	r.DevName = devName
	r.Created = time.Now()
	r.Updated = time.Now()
	r.Sync = sync
	r.Cache = false
	r.Chunk = 256

	// init RaidDisks
	dataDisks, err := GetDisksByArgv(map[string]interface{}{"location": raid})
	if err != nil {
		util.AddLog(err)
		return
	}
	r.RdsNr = int64(len(dataDisks))

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
	r.DbRaid.createCache() // Create Cache

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
		if r.GetHealth() != HEALTH_FAILED {
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

	r.DbRaid.Save(map[string]interface{}{"Deleted": true})
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
		case "Name":
			r.Name = v.(string)
		case "Health":
			r.Health = v.(string)
		case "Chunk":
			r.Chunk = v.(int64)
		case "Cap":
			r.Cap = v.(int64)
		case "UsedCap":
			r.UsedCap = v.(int64)
		case "DevName":
			r.DevName = v.(string)
		case "OdevName":
			r.OdevName = v.(string)
		case "UnplugSeq":
			r.UnplugSeq = v.(int64)
		case "Deleted":
			r.Deleted = v.(bool)
		}
	}
}

// Get raid disks
func (r *DbRaid) RaidDisks() (disks []DbDisk) {
	o := orm.NewOrm()

	// get foreign keys
	if _, err := o.LoadRelated(r, "Disks"); err != nil {
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
	if _, err := o.LoadRelated(r, "Disks"); err != nil {
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
		cmd := fmt.Sprintf("mdadm --grow --bitmap=%s --bitmap-chunk=%sM %s", bitmap, strconv.Itoa(bitmapSize), r.DevPath())
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
	cacheEnabled, _ := beego.AppConfig.Bool("cache_enable")
	raidcachememsize, _ := beego.AppConfig.Int("raid_cache_mem_size")
	if cacheEnabled {
		//var cmd []string
		//cmd = append(cmd, "--getsz", r.OdevPath())
		cmd := fmt.Sprintf("blockdev --getsz %s", r.OdevPath())
		//size, err := util.Execute("blockdev", cmd, true)
		size, err := util.ExecuteByStr(cmd, true)
		fmt.Printf("size:\n%+v %+v", size, err)
		if err != nil {
			//		return
		}
		odev_name := "c" + r.DevName
		rule := fmt.Sprintf("0 %s cache %s %s %s %s", strings.TrimSpace(size), r.OdevPath(), strconv.Itoa(raidcachememsize*raidcachememsize), strconv.FormatInt(r.Chunk*r.Chunk, 10), strconv.FormatInt(r.RdsNr-1, 10)) //TODO
		dmcreate(odev_name, rule)
		r.Save(map[string]interface{}{"OdevName": odev_name})
	}
}

// Raids Lookup
func GetRaidsByArgv(items map[string]interface{}) (r Raid, err error) {
	o := orm.NewOrm()

	for k, v := range items {
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

	return
}

// func AddRaids
// Update extents
func (r *DbRaid) updateExtents() (err error) {
	if r.GetHealth() == HEALTH_DEGRADED || r.GetHealth() == HEALTH_NORMAL {
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
	var raid_disk_paths []string
	for _, d := range r.RaidDisks {
		raid_disk_paths = append(raid_disk_paths, d.DevPath())
	}

	mdadmUuid := util.Format(util.MDADM_FORMAT, r.Id)

	var sync, cmd string

	if r.Sync {
		sync = ""
	} else {
		sync = "--assume-clean"
	}

	bitmapFile := beego.AppConfig.String("bitmapfile")
	bitmap := bitmapFile + r.Id + ".bitmap"
	homehost := beego.AppConfig.String("raid_homehost")

	//TODO chunk
	//TODO --bitmap-chunk
	//TODO --layout
	level := strconv.FormatInt(r.Level, 10)
	count := strconv.Itoa(len(raid_disk_paths))
	if r.Level == 0 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"%s\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run --force -q --name=\"%s\"",
			r.DevPath(), homehost, mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), r.Name)
	} else if r.Level == 1 || r.Level == 10 {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"%s\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=%s --bitmap-chunk=16M",
			r.DevPath(), homehost, mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)

	} else {
		cmd = fmt.Sprintf("mdadm --create %s --homehost=\"%s\" --uuid=\"%s\" --level=%s "+
			"--chunk=256 --raid-disks=%s %s --run %s --force -q --name=\"%s\" --bitmap=%s --bitmap-chunk=16M --layout=left-symmetric",
			r.DevPath(), homehost, mdadmUuid, level, count, strings.Join(raid_disk_paths, " "), sync, r.Name, bitmap)
	}

	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	if err = r._clean_existed_partition(); err != nil {
		util.AddLog(err)
		return
	}

	var rr RaidRecovery
	var raidUuids, spareUuids []string

	for _, d := range r.RaidDisks {
		raidUuids = append(raidUuids, d.Id)
	}

	for _, d := range r.SpareDisks {
		spareUuids = append(spareUuids, d.Id)
	}
	rr.Save(map[string]interface{}{"uuid": r.Id, "name": r.Name, "level": r.Level, "chunk_kb": r.Chunk, "raid_disks_nr": len(r.RaidDisks), "raid_disks": strings.Join(raidUuids, ","), "spare_disks": strings.Join(spareUuids, ","), "spare_disks_nr": len(r.SpareDisks)})

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
func (r *DbRaid) GetHealth() string {
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
	//cacheEnabled, _ := beego.AppConfig.Bool("cache_enable")
	if false {
		return "/dev/mapper/c" + r.DevName
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

// Checking by function AddRaids
func CheckRaidsByArgv(items map[string]interface{}) (err error) {
	o := orm.NewOrm()

	for k, v := range items {

		switch k {
		case "name":
			if exist := o.QueryTable(new(DbRaid)).Filter(k, v).Exist(); exist {
				err = fmt.Errorf(v.(string) + " has existed")
				return
			}
		case "level":
			if _, ok := LEVEL[v.(int64)]; !ok {
				err = fmt.Errorf("invalid param level=" + strconv.FormatInt(v.(int64), 10))
				return
			}
		case "raid_disks", "spare_disks":
			disks := strings.FieldsFunc(v.(string), func(c rune) bool { return c == ',' })
			level := items["level"].(int64)
			count := len(disks)

			if count == 0 && k == "spare_disks" {
				continue
			}
			for _, disk := range disks {
				if exist := o.QueryTable(new(DbDisk)).Filter("location", disk).Exist(); !exist {
					err = fmt.Errorf(k + " location=" + disk + " has not exist")
					return
				} else if exist := o.QueryTable(new(DbDisk)).Filter("health", HEALTH_NORMAL).Filter("role", ROLE_UNUSED).Filter("location", disk).Exist(); !exist {
					err = fmt.Errorf(k + " " + disk + " has used")
					return
				}
			}

			if k == "spare_disks" {
				continue
			}

			if count == 0 {
				err = fmt.Errorf("0 disks are not enough to create level " + strconv.FormatInt(level, 10) + " raid")
				return
			}

			if level == 1 {
				if count > 2 {
					err = fmt.Errorf("level 1 raid only support 2 disks")
					return
				} else if count < 2 {
					err = fmt.Errorf(strconv.Itoa(count) + " disks are not enough to create level 1 raid")
					return
				}
			}
			if level == 5 && count < 3 {
				err = fmt.Errorf(strconv.Itoa(count) + " disks are not enough to create level 5 raid")
				return
			}
		case "rebuild_priority":
			if _, ok := REBUILD_PRIORITY[v.(string)]; !ok {
				err = fmt.Errorf("invalid param rebuild=" + v.(string))
				return
			}
		case "chunk_kb":
			if v.(int64)%64 != 0 {
				err = fmt.Errorf("invalid param chunk=" + strconv.FormatInt(v.(int64), 10))
				return
			}
		}
	}
	return
}

// Get raid's dev name
func nextDevName() (string, error) {
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

	fmt.Println(rules)
	cmd := fmt.Sprintf("touch %s", tmpfile)
	util.WriteFile(tmpfile, rules+"\n")
	//create file

	util.ExecuteByStr(cmd, true)

	cmd = fmt.Sprintf("dmsetup create %s %s", name, tmpfile)
	util.ExecuteByStr(cmd, true)

	dm_path := "/dev/mapper/" + name
	ensureExist(dm_path)

	//os.Remove(tmpfile)
}

/***********************scan raid**********************/

// Find back disk
func (r *DbRaid) findBackDisks() ([]DbDisk, int64) {
	disks := make([]DbDisk, 0)
	unplugSeq := r.UnplugSeq

	if unplugSeq > 0 {
		var bydisks ByUnplugSeq
		for _, disk := range r.RaidDisks() {
			if disk.Role == ROLE_DATA_SPARE && disk.Online() && disk.Health != HEALTH_FAILED && !disk.Link {
				bydisks = append(bydisks, disk)
			}
		}
		sort.Sort(bydisks)
		for _, disk := range bydisks {
			if disk.UnplugSeq > 0 && unplugSeq == disk.UnplugSeq {
				disks = append(disks, disk)
				unplugSeq = unplugSeq - 1
			} else {
				break
			}
		}
	}

	return disks, unplugSeq
}

// if can online
func (r *DbRaid) canOnline(disks []DbDisk) bool {
	disk_devs := make([]string, 0)
	for _, disk := range disks {
		disk_devs = append(disk_devs, disk.DevPath())
	}
	cmd := "mdadm --assemble --force --run /dev/md128 " + strings.Join(disk_devs, " ")
	_, err := util.ExecuteByStr(cmd, true)
	defer func() {
		cmd := "mdadm --stop /dev/md128"
		util.ExecuteByStr(cmd, true)
	}()

	if err != nil {
		return false
	}
	return true
}

// func mdadmCreate TODO change method
func (r *DbRaid) activeRebuildPriority() {
	//TODO min

	if len(r.RaidDisks()) < 12 {
		cmd := fmt.Sprintf("echo 20480 > /sys/block/%s/md/stripe_cache_size", r.DevName)
		if _, err := util.ExecuteByStr(cmd, true); err != nil {
			return
		}
	}
	cmd := fmt.Sprintf("echo 0 > /sys/block/%s/md/preread_bypass_threshold", r.DevName)
	if _, err := util.ExecuteByStr(cmd, true); err != nil {
		return
	}

	return
}

// recreate raid when reboot
func (r *DbRaid) recreateWhenReboot() bool {
	recoveries, _ := SelectRaidRecovery(map[string]interface{}{"Id": r.Id}, []string{"create_at"}, "desc")
	if len(recoveries) == 0 {
		return false
	}

	devname, err := nextDevName()
	if err != nil {
		return false
	}

	r.DevName = devname
	mdadmUUID := util.Format(util.MDADM_FORMAT, r.Id)
	bitmap := beego.AppConfig.String("bitmapfile") + r.Id + ".bitmap"
	if _, err := os.Stat(bitmap); err == nil {
		os.Remove(bitmap)
	}

	diskDevs := make([]string, 0)
	for _, d := range strings.Split(recoveries[0].RaidDisks, ",") {
		ds, _ := GetDisksByArgv(map[string]interface{}{"Id": d})
		if len(ds) != 0 {
			diskDevs = append(diskDevs, ds[0].DevPath())
		}
	}

	if len(diskDevs) != len(r.RaidDisks()) {
		return false
	}

	level := strconv.FormatInt(r.Level, 10)
	count := strconv.Itoa(len(r.RaidDisks()))
	cmd := fmt.Sprintf("mdadm --create %s --homehost=\"speedio\" --uuid=\"%s\" --level=%s --chunk=256 "+
		"--raid-disks=%s %s --run --assume-clean --force -q --name=\"%s\" --bitmap=%s "+
		"--bitmap-chunk=16M --layout=left-symmetric",
		r.DevPath(), mdadmUUID, level, count, strings.Join(diskDevs, " "), r.Name, bitmap)

	if _, err = util.ExecuteByStr(cmd, true); err != nil {
		return false
	}

	var stat os.FileInfo
	stat, err = os.Stat(r.DevPath())
	if err != nil {
		return false
	}
	st, _ := stat.Sys().(*syscall.Stat_t)
	r.Save(map[string]interface{}{"DevNo": st.Rdev, "OdevName": r.DevName})

	if r.Level == 5 || r.Level == 6 {
		r.activeRebuildPriority()
	}

	return true
}

// assemble disks into raid
func (r *DbRaid) _assemble() {
	defer func() {
		cmd := "pvscan"
		util.ExecuteByStr(cmd, true)
	}()

	diskDevs := make([]string, 0)
	raidDisks, lastRDisks := make([]DbDisk, 0), make([]DbDisk, 0)

	for _, disk := range r.RaidDisks() {
		if disk.Role == ROLE_DATA && disk.Online() && disk.Health != HEALTH_FAILED {
			raidDisks = append(raidDisks, disk)
		}
	}
	for _, disk := range r.RaidDisks() {
		if disk.Role == ROLE_DATA_SPARE &&
			disk.Online() &&
			disk.Health != HEALTH_FAILED &&
			!disk.Link &&
			disk.UnplugSeq == 0 {
			lastRDisks = append(lastRDisks, disk)
		}
	}

	if len(lastRDisks) == 1 {
		lastRDisks[0].Save(map[string]interface{}{"UnplugSeq": 1})
		r.Save(map[string]interface{}{"UnplugSeq": 1})
	}
	disks, unplugSeq := r.findBackDisks()
	if unplugSeq == 0 && len(disks) > 0 {
		disks = disks[:len(disks)-1]
		unplugSeq = 1
	}
	raidDisks = append(raidDisks, disks...)
	r.DevName, _ = nextDevName()
	if !r.canOnline(raidDisks) {
		return
	}

	for _, disk := range r.RaidDisks() {
		disk.Save(map[string]interface{}{"link": true, "role": ROLE_DATA, "unplug_seq": 0})
		diskDevs = append(diskDevs, disk.DevPath())
	}

	recreated := false
	if len(diskDevs) == len(r.RaidDisks()) && (r.Level == 5 || r.Level == 6) && r.GetHealth() == HEALTH_NORMAL {
		recreated = r.recreateWhenReboot()
	}
	if recreated {
		return
	}

	cmd := ""
	if r.Level == 5 {
		bitmap := filepath.Join("bitmap", fmt.Sprintf("%s.bitmap", r.Id))
		_, err := os.Stat(bitmap)
		if err == nil {
			cmd = fmt.Sprintf("mdadm --assemble --force --run --bitmap=%s %s %s", bitmap, r.DevPath(), strings.Join(diskDevs, " "))
		}
	}
	if cmd == "" {
		cmd = fmt.Sprintf("mdadm --assemble --force --run %s %s", r.DevPath(), strings.Join(diskDevs, " "))
	}

	_, err := util.ExecuteByStr(cmd, true)
	if err != nil {
		util.AddLog(err)
		return
	}

	time.Sleep(1 * time.Second)
	if r.Online() && r.GetHealth() == HEALTH_NORMAL && !r.hasBitmap() {
		r.createBitmap()
	}
	/*var stat os.FileInfo
	stat, err = os.Stat(r.DevPath())
	if err != nil {
		return
	}
	st, _ := stat.Sys().(*syscall.Stat_t)
	r.Save(map[string]interface{}{"DevNo": st.Rdev, "UnplugSeq": unplugSeq, "OdevName": r.DevName}) TODO*/
	r.Save(map[string]interface{}{"UnplugSeq": unplugSeq, "OdevName": r.DevName})

	if r.Level == 5 || r.Level == 6 {
		r.activeRebuildPriority()
	}
}

// assemble disks into raid
func (r *DbRaid) assemble() {
	r._assemble()
	if r.Online() {
		r.createCache()
		//TODO ueventd
	}
}

// Scan raid
func (r *DbRaid) Scan() {
	for _, disk := range r.RaidDisks() {
		if !disk.Online() {
			r.Save(map[string]interface{}{"Role": ROLE_DATA_SPARE})
		}
	}

	r.Save(map[string]interface{}{"DevName": "", "OdevName": ""})
	r.assemble()

	return
}

func RaidAll() (rds []DbRaid, err error) {
	o := orm.NewOrm()
	if _, err = o.QueryTable(new(DbRaid)).All(&rds); err != nil {
		util.AddLog(err)
		return
	}
	return
}

func ScanAllRaids() error {
	rds, err := RaidAll()
	if err != nil {
		return err
	}

	for _, rd := range rds {
		rd.Scan()
	}

	cmd := "pvscan"
	_, err = util.ExecuteByStr(cmd, true)
	if err != nil {
		return err
	}

	return nil
}

/*
	err = fmt.Errorf("")
		return
*/
