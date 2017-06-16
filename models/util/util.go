package util

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego/logs"
	"os"
	"os/exec"
	"runtime"
)

// Uuid created
func Urandom() string {
	f, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	l := make([]byte, 16)
	f.Read(l)
	f.Close()
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", l[0:4], l[4:6], l[6:8], l[8:10], l[10:])
	return uuid
}

// Cmd
func ExecuteByStr(cmdArgs string) (output string, err error) {
	AddLog(cmdArgs)

	cmd := exec.Command("/bin/sh", "-c", cmdArgs)

	// Stdout buffer
	w := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stderr = w
	cmd.Stdout = w
	// Execute command
	err = cmd.Run() // will wait for command to return
	if err != nil {
		AddLog(err)
		return
	}

	return string(w.Bytes()), nil
}

// Cmd
func Execute(name string, cmdArgs []string) (output string, err error) {
	cmd := exec.Command(name, cmdArgs...)

	// Stdout buffer
	w := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stderr = w
	cmd.Stdout = w
	// Execute command
	err = cmd.Run() // will wait for command to return

	return string(w.Bytes()), nil
}

// Logs
func AddLog(err interface{}, v ...interface{}) {
	if _, ok := err.(error); ok {
		pc, _, line, _ := runtime.Caller(1)
		logs.Error("[Server] ", runtime.FuncForPC(pc).Name(), line, v, err)
	} else {
		logs.Info("[Server] ", err)
	}
}
