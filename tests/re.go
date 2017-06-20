package main

import (
	"fmt"
	"regexp"
	"speedio/models/util"
	"strings"
)

func main() {

	fmt.Println(util.CheckSystemRaid1())

	/*cmd := make([]string, 0)
	cmd = append(cmd, "--detail", "/dev/md0")
	output, _ := util.Execute("mdadm", cmd)*/

	//	re := regexp.MustCompile("State :\\s+([^\n]+)")

	/*status := strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' })
	sMap := make(map[string]bool, 0)
	for _, s := range status {
		sMap[s] = true
	}

	fmt.Println("segs:", segs)
	fmt.Println("segs:", len(segs))
	fmt.Println("segs:", strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' }))*/
	//fmt.Println(segs[0], ",", segs[1], ",", segs[2], ",", segs[3])
}

func re1() {
	cmd := fmt.Sprintf("ls -l /sys/block/ |grep 'ata' |grep -v '%s'", "sda")
	o, _ := util.ExecuteByStr(cmd, false)
	o = strings.TrimSpace(o)
	fmt.Println(len(strings.Split(o, "\n")))
	//re := regexp.MustCompile(`ata(\d+)`)
	re := regexp.MustCompile(`target(\d+):(\d+)`)
	segs := re.FindAllString(o, -1)
	fmt.Println("segs:", segs)
	fmt.Printf("\nsegs:%s\n", re.Find([]byte(o)))
	fmt.Printf("\nsegs:%s\n", re.FindSubmatch([]byte(o)))

	id := string(re.FindSubmatch([]byte(o))[1])
	fmt.Println(id)
	fmt.Println(regexp.MatchString(`ata(\d+)`, o))
	fmt.Println(re.MatchString(o))
}
