package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	_, err := os.Stat("/dev/sdaf")
	fmt.Println(err)
	//	a, err := os.Stat("/dev/md0")
	if !os.IsExist(err) {
		// path/to/whatever does not exist
		fmt.Println(err)
	}
	//fmt.Printf("%+v", a.IsDir(), err)
	fmt.Println(_next_dev_name())
	//fmt.Println("\n", a, b, "\n", c, d, "\n", filepath.SplitList("/dev/md?*"))
	//fmt.Println(_next_dev_name())
}

func _next_dev_name() (string, error) {
	nr, err := filepath.Glob("/dev/md?*")
	if err != nil {
		return "md0", err
	}

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
