package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "\n Hello, World\n "
	fmt.Printf("%d %q\n", len(s), s)
	t := strings.TrimSpace(s)
	fmt.Printf("%d %q\n", len(t), t)
	fmt.Printf("[%q]", strings.Trim(" Achtung  ", " "))
}
