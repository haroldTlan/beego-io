package main

import (
	"fmt"
	"regexp"
	"speedio/models/util"
	"strings"
)

func main() {

	cmd := make([]string, 0)
	cmd = append(cmd, "--detail", "/dev/md0")
	output, _ := util.Execute("mdadm", cmd)

	re := regexp.MustCompile("State :\\s+([^\n]+)")
	segs := re.FindAllString(output, -1)

	status := strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' })
	sMap := make(map[string]bool, 0)
	for _, s := range status {
		sMap[s] = true
	}

	fmt.Println("segs:", segs)
	fmt.Println("segs:", len(segs))
	fmt.Println("segs:", strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' }))
	//fmt.Println(segs[0], ",", segs[1], ",", segs[2], ",", segs[3])
}
