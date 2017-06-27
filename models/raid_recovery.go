package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"speedio/util"
)

type RaidRecovery struct {
	Id           int       `orm:"column(id);auto"                  json:"id"`
	Created      time.Time `orm:"column(created_at);type(datetime)"   json:"created_at"`
	Updated      time.Time `orm:"column(updated_at);type(datetime)"   json:"updated_at"`
	Uuid         string    `orm:"column(uuid);size(255)"              json:"uuid"`
	Name         string    `orm:"column(name);size(255)"              json:"name"`
	Level        int64     `orm:"column(level)"                       json:"level"`
	Chunk        int64     `orm:"column(chunk_kb)"                    json:"chunk_kb"`
	RaidDisksNr  int64     `orm:"column(raid_disks_nr)"               json:"raid_disks_nr"`
	RaidDisks    string    `orm:"column(raid_disks);type(text)"       json:"raid_disks"`
	SpareDisksNr int64     `orm:"column(spare_disks_nr)"              json:"spare_disks_nr"`
	SpareDisks   string    `orm:"column(spare_disks);type(text)"      json:"spare_disks"`
}

func (rr *RaidRecovery) TableName() string {
	return "raid_recoveries"
}

func init() {
	orm.RegisterModel(new(RaidRecovery))
}

// GET
// Get Raid_recoveries
// Can use sort
func SelectRaidRecovery(item map[string]interface{}, sortby []string, order string) (rs []RaidRecovery, err error) {
	o := orm.NewOrm()

	qs := o.QueryTable(new(RaidRecovery))

	for k, v := range item {
		// check k or v
		qs = qs.Filter(k, v)
	}

	var sortFields []string
	if len(sortby) != 0 {
		orderby := ""
		for _, v := range sortby {
			if order == "desc" {
				orderby = "-" + v
			} else if order == "asc" {
				orderby = v
			} else {
				err = fmt.Errorf("Error: Invalid order. Must be either [asc|desc]")
				util.AddLog(err)
				return
			}
		}
		sortFields = append(sortFields, orderby)

	}

	qs = qs.OrderBy(sortFields...)
	if _, err = qs.All(&rs); err != nil {
		util.AddLog(err)
		return
	}

	return
}

// POST
// Insert raidRecoveries infos
func (rr *RaidRecovery) Save(items map[string]interface{}) (err error) {
	o := orm.NewOrm()

	// checking value TODO
	for k, v := range items {
		switch k {
		case "uuid":
			rr.Uuid = v.(string)
		case "name":
			rr.Name = v.(string)
		case "level":
			rr.Level = v.(int64)
		case "chunk_kb":
			rr.Chunk = v.(int64)
		case "raid_disks_nr":
			rr.RaidDisksNr = int64(v.(int))
		case "raid_disks":
			rr.RaidDisks = v.(string)
		case "spare_disks_nr":
			rr.SpareDisksNr = int64(v.(int))
		case "spare_disks":
			rr.SpareDisks = v.(string)
		}
	}

	rr.Id = 0
	rr.Created = time.Now()
	rr.Updated = time.Now()
	fmt.Printf("9%+v\n", rr)
	if _, err = o.Insert(rr); err != nil {
		util.AddLog(err)
		return
	}

	return
}

func RecordRaidRecovery(raid DbRaid) {

	resolveSeq := func(seq, refer []string) []string {

		fmt.Printf("\nseq:  %+v\n\n", seq)
		fmt.Printf("\n0refer:  %+v\n\n", refer)
		if len(seq) == 0 || len(refer) == 0 {
			return []string{}
		}

		refMap, seqMap := make(map[string]int), make(map[string]int)

		for i, id := range refer {
			refMap[id] = i
		}

		for i, id := range seq {
			seqMap[id] = i
		}

		seqDiff := make([]int, 0)
		for i, id := range seq {
			if _, ok := refMap[id]; !ok {
				seqDiff = append(seqDiff, i)
			}
		}
		refDiff := make([]int, 0)
		for i, id := range refer {
			if _, ok := seqMap[id]; !ok {
				refDiff = append(refDiff, i)
			}
		}

		fmt.Printf("\n1refer:  %+v\n\n", refer)
		if len(refDiff) == 0 || len(seqDiff) == 0 {
			return []string{}
		}
		refer[refDiff[0]] = seq[seqDiff[0]]

		fmt.Printf("\n2refer:  %+v\n\n", refer)
		return refer
	}

	sortby := []string{"created_at"}
	recoveries, err := SelectRaidRecovery(map[string]interface{}{"uuid": raid.Id}, sortby, "desc")
	if err != nil {
		util.AddLog(err)
		return
	}

	var uuids, recovery []string
	var rr RaidRecovery
	for _, d := range raid.RaidDisks() {
		uuids = append(uuids, d.Id)
	}

	if len(recoveries) > 0 {
		recovery = strings.FieldsFunc(recoveries[0].RaidDisks, func(c rune) bool { return c == ',' })
		rr = recoveries[0]
	}

	seq := resolveSeq(uuids, recovery)

	fmt.Printf("2%+v\n\n", seq)
	if len(seq) > 0 {
		util.AddLog(fmt.Sprintf("resolve correct created raid disks: %s", seq))
	} else {
		seq = recovery
	}

	var spareDisks []string
	for _, d := range raid.SpareDisks() {
		spareDisks = append(spareDisks, d.Id)
	}
	spareDisk := strings.Join(spareDisks, ",")

	rr.Save(map[string]interface{}{"uuid": raid.Id, "name": raid.Name, "level": raid.Level, "chunk_kb": raid.Chunk, "raid_disks_nr": len(seq), "raid_disks": strings.Join(seq, ","), "spare_disks": spareDisk, "spare_disks_nr": len(raid.SpareDisks())})
}
