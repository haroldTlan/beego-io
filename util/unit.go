package util

import (
	"fmt"
	"math"
)

type Unit struct {
	Byte float64
	KB   float64
	MB   float64
	GB   float64
	TB   float64
}

var UNIT = map[string]float64{
	"B": math.Pow(1024, 0), "b": math.Pow(1024, 0),
	"K": math.Pow(1024, 1), "k": math.Pow(1024, 1),
	"M": math.Pow(1024, 2), "m": math.Pow(1024, 2),
	"G": math.Pow(1024, 3), "g": math.Pow(1024, 3),
	"T": math.Pow(1024, 4), "t": math.Pow(1024, 4),
	"P": math.Pow(1024, 5), "p": math.Pow(1024, 5),
	"KB": math.Pow(1024, 1), "MB": math.Pow(1024, 2),
	"GB": math.Pow(1024, 3), "TB": math.Pow(1024, 4),
	"PB": math.Pow(1024, 5), "KiB": math.Pow(1024, 1),
	"MiB": math.Pow(1024, 2), "GiB": math.Pow(1024, 3),
	"TiB": math.Pow(1024, 4), "PiB": math.Pow(1024, 5),
	"BYTE": math.Pow(1024, 0), "S": 512,
	"Page": 4096, "SECTOR": 512,
}

func UnitFormat(val int, unit string) (u Unit, err error) {
	if _, ok := UNIT[unit]; !ok {
		err = fmt.Errorf("FormatError")
		AddLog(err)
		return
	}

	u.Byte = float64(val) * UNIT[unit]
	u.KB = float64(val) * UNIT[unit] / UNIT["KB"]
	u.MB = float64(val) * UNIT[unit] / UNIT["MB"]
	u.GB = float64(val) * UNIT[unit] / UNIT["GB"]
	u.TB = float64(val) * UNIT[unit] / UNIT["TB"]
	return
}

func (u *Unit) ToInt64() int64 {
	return int64(u.Byte)
}

func (u *Unit) ToInt(str string) int {
	return int(u.Byte)
}
