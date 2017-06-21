package main

import (
	"fmt"
	"os"
	"speedio/models/util"
)

func main() {
	fmt.Println("vim-go")
	rules := "0 23 cache cmd0 123 123 123123"
	util.WriteFile("/tmp/a.conf", rules)
	os.Remove("/tmp/a.conf")

}
