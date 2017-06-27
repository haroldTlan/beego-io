package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"speedio/models/lm"
	"speedio/util"
)

func Inits() (res string, err error) {
	o := orm.NewOrm()
	// Delete Raids
	cmd := "mdadm --stop /dev/md*"
	util.ExecuteByStr(cmd, true)

	// Delete sqlite
	if _, err = o.QueryTable(new(DbDisk)).Filter("uuid__isnull", true).Delete(); err != nil {
		return
	}

	if _, err = o.QueryTable(new(DbRaid)).Filter("uuid__isnull", true).Delete(); err != nil {
		return
	}

	// Init disks
	var m lm.LocationMapping
	m.Init()

	for _, dev := range m.Mapping {
		cmd = fmt.Sprintf("dd if=/dev/zero of=/dev/%s bs=1M count=12", dev)
		if _, err = util.ExecuteByStr(cmd, true); err != nil {
			return
		}
	}

	return
}
