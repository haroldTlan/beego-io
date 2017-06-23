package main

import (
	"fmt"
	"regexp"
	"speedio/models/util"
	"strings"
)

func main() {
	cmd := fmt.Sprintf("hdparm -I /dev/%s", "sdd")
	o, _ := util.ExecuteByStr(cmd, true)

	m := regexp.MustCompile(`Model Number:\s*(.+)`)
	mStr := m.FindSubmatch([]byte(o))
	fmt.Println(len(mStr))
	fmt.Println(string(mStr[1]))
}
func cmd1() {
	cmd := make([]string, 0)
	cmd = append(cmd, "/dev/mapper/cmd0", "-o", "pv_pe_alloc_count,pv_pe_count")
	output, _ := util.Execute("pvs", cmd)

	vgsMap := make(map[string]bool, 0)
	o := strings.Fields(output)
	for _, vgs := range strings.Fields(output) {
		vgsMap[vgs] = true
	}
	fmt.Println(o)
}
