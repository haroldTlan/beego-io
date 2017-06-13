package util

import (
	"bytes"
	_ "fmt"
	"os/exec"
)

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
