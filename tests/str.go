package main

import (
	"fmt"
	"strings"
)

func main() {
	//b := "1.1.2,1.1.3,1.14,1.1.5"
	b := "1.1.1"
	c := ""
	fmt.Println(`vim` + b + `-go`)
	fmt.Println("\"vim-go\"")
	fmt.Println(len(strings.FieldsFunc(c, func(c rune) bool { return c == ',' })))
	fmt.Println(len(strings.Split(b, ",")))
	a := strings.Fields(" ")
	fmt.Println("vim-go", len(a))

}
