package main

import (
	"fmt"
	"strings"

	"speedio/models/util"
)

func main() {
	//segs = strings.TrimSpace(segs)
	//b := strings.Fields(segs)
	//fmt.Println(len(b))
	uuid := "b8c463dc-3a38-2471-00a6-5252106eb46c"
	fmt.Println(util.Validate(uuid))
}

func format() {
	uuid := "2aae1085-f524-4de9-8db0-aa81d87dfcc4"
	fmt.Println(util.Format(util.MDADM_FORMAT, uuid))
}

func uuid() {
	uuid := "2aae1085-f524-4de9-8db0-aa81d87dfcc4"
	l := strings.Join(strings.Split(uuid, "-"), "")
	a := fmt.Sprintf("%s-%s-%s-%s", l[0:8], l[8:16], l[16:24], l[24:])
	fmt.Println(a)
}
