package main

import (
	"fmt"
	"speedio/models/util"
	"strings"
)

func main() {
	cmd := make([]string, 0)
	cmd = append(cmd, "/dev/mapper/cmd0", "-o", "pv_pe_alloc_count,pv_pe_count")
	output, _ := util.Execute("pvs", cmd)

	vgsMap := make(map[string]bool, 0)
	o := strings.Fields(output)
	fmt.Println(o[len(o)-1], o[len(o)-2])
	fmt.Printf("%+v", strings.Split(output, "\n")[1])
	for _, vgs := range strings.Fields(output) {
		vgsMap[vgs] = true
	}
	fmt.Printf("%+v", vgsMap)
}
