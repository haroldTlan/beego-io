package main

import (
	"fmt"
	"reflect"
)

type A struct {
	B string
	C int
	D int64
	E float64
}

func main() {
	fmt.Println("vim-go")
	var a A
	a.B = "asd"
	a.C = 2
	a.D = 33
	a.E = 2.98

	t := reflect.TypeOf(a)
	PrintType(t)

	if field, ok := t.FieldByName("B"); ok {
		fmt.Println("FieldByName(ptr)   :", field.Name)
	}
}

func PrintType(t reflect.Type) {
	fmt.Println("String             :", t.String())     // 类型字符串
	fmt.Println("Name               :", t.Name())       // 类型名称
	fmt.Println("PkgPath            :", t.PkgPath())    // 所在包名称
	fmt.Println("Kind               :", t.Kind())       // 所属分类
	fmt.Println("Size               :", t.Size())       // 内存大小
	fmt.Println("Align              :", t.Align())      // 字节对齐
	fmt.Println("FieldAlign         :", t.FieldAlign()) // 字段对齐
	fmt.Println("NumMethod          :", t.NumMethod())  // 方法数量

	fmt.Println("=== 结构体 ===")
	fmt.Println("NumField           :", t.NumField()) // 字段数量
	if t.NumField() > 0 {
		var i, j int
		// 遍历结构体字段
		for i = 0; i < t.NumField()-1; i++ {
			field := t.Field(i) // 获取结构体字段
			fmt.Printf("    ├ %v\n", field.Name)
			// 遍历嵌套结构体字段
			if field.Type.Kind() == reflect.Struct && field.Type.NumField() > 0 {
				for j = 0; j < field.Type.NumField()-1; j++ {
					subfield := t.FieldByIndex([]int{i, j}) // 获取嵌套结构体字段
					fmt.Printf("    │    ├ %v\n", subfield.Name)
				}
				subfield := t.FieldByIndex([]int{i, j}) // 获取嵌套结构体字段
				fmt.Printf("    │    └ %%v\n666", subfield.Name)
			}
		}
		field := t.Field(i) // 获取结构体字段
		fmt.Printf("    └ %v\n???", field.Name)
	}
}
