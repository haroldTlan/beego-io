package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

func main() {

	q := map[uint64]string{uint64(22): "sd", uint64(12): "er", uint64(19): "dfg", uint64(16): "pl"}
	for _, i := range q {
		fmt.Println(i)
	}
	sort.Stable(q)

	//	a, err := os.Stat("/dev/md0")
	//if os.IsExist(err) {
	// path/to/whatever does not exist
	//	fmt.Println(err)
	//}
	//fmt.Printf("%+v", a.IsDir(), err)
	//fmt.Println(_next_dev_name())
	//fmt.Println("\n", a, b, "\n", c, d, "\n", filepath.SplitList("/dev/md?*"))
	//fmt.Println(_next_dev_name())
}

func st_rdev() {
	dev, _ := os.Stat("/dev/sda")
	sys, ok := dev.Sys().(*syscall.Stat_t)
	if !ok {
		os.Exit(1)
	}
	major := sys.Rdev / 256
	minor := sys.Rdev % 256
	devNumStr := fmt.Sprintf("%d:%d", major, minor)
	fmt.Printf("get dev mapper [%s] [%s]", dev.Name, devNumStr)

	stat := syscall.Stat_t{}
	_ = syscall.Stat("/dev/sda", &stat)
	fmt.Println("Major:", uint64(stat.Rdev/256), "Minor:", uint64(stat.Rdev%256))

}

func _next_dev_name() (string, error) {
	nr, err := filepath.Glob(`/dev/sda*`)
	if err != nil {
		return "md0", err
	}
	for _, i := range nr {
		fmt.Println(filepath.Base(i))
	}
	fmt.Println(nr)
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
