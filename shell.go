package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunShellScript executes shell script with supplied arguments.
func RunShellScript(script string, args ...string) {
	cmd := exec.Command(script, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if nil != err {
		fmt.Println(stderr.String())
		return
	}
	fmt.Println(stdout.String())
}
