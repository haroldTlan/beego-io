package main

import (
	"fmt"
	"strings"
)

func main() {
	segs := "asd  asdasdasd                    "
	//segs = strings.TrimSpace(segs)
	b := strings.Fields(segs)
	fmt.Println(len(b))

}

func uuid() {
	uuid := "2aae1085-f524-4de9-8db0-aa81d87dfcc4"
	l := strings.Join(strings.Split(uuid, "-"), "")
	a := fmt.Sprintf("%s-%s-%s-%s", l[0:8], l[8:16], l[16:24], l[24:])
	fmt.Println(a)
}
