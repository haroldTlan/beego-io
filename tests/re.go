package main

import (
	"fmt"
	"regexp"
	"speedio/models/util"
	"strings"
)

func main() {

	uuid := "b72e1b23-2c3c-4049-89bd-c0fb101fa05e"
	VALID_CHAR := "[0-9|a-f|A-F]"
	SPEEDIO_FORMAT := "8-4-4-4-12"
	fmts := SPEEDIO_FORMAT

	m := regexp.MustCompile(`\d+(.)`)
	sep := m.FindSubmatch([]byte(fmts))[1]
	var char []string
	for _, nr := range strings.Split(fmts, string(sep)) {
		fmt.Println(nr)
		c := fmt.Sprintf("%s{%s}", VALID_CHAR, nr)
		char = append(char, c)
	}
	pattern := strings.Join(char, string(sep))
	fmt.Println(string(sep), char)
	fmt.Printf("%+v", pattern)
	fmt.Println(regexp.MatchString(pattern, uuid))
	//fmt.Println(util.CheckSystemRaid1())

	/*	cmd := make([]string, 0)
		cmd = append(cmd, "--detail", "/dev/md0")
		output, _ := util.Execute("mdadm", cmd)*/

	//fmt.Println(segs[0], ",", segs[1], ",", segs[2], ",", segs[3])
}

func re2() {
	cmd := fmt.Sprintf("mdadm --detail /dev/md0")
	output, err := util.ExecuteByStr(cmd, true)
	fmt.Println("segs:", output, err)
	re := regexp.MustCompile(`State :\s+([^\n]+)`)
	segs := re.FindAllString(output, -1)

	status := strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' })
	sMap := make(map[string]bool, 0)
	for _, s := range status {
		sMap[s] = true
	}

	fmt.Println("segs:", segs)
	fmt.Println("segs:", len(segs))
	fmt.Println("segs:", strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' }))
	fmt.Println("segs:", len(strings.FieldsFunc(segs[0], func(c rune) bool { return c == ':' })))
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
