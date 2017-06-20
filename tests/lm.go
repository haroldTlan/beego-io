package main

import (
	"fmt"

	"github.com/astaxie/beego"
	"speedio/models/lm"
)

var (
	company = beego.AppConfig.String("company")
	sn, _   = beego.AppConfig.Int("slotNum")
)

func main() {
	//fmt.Println("vim-go", a)
	fmt.Printf("%+v\n", lm.RealToLogicMapping(company, sn))
}
