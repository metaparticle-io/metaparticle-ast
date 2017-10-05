package compiler

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func executeCommand(cmd []string, dryrun bool) error {
	if dryrun {
		fmt.Printf("Would execute: %v\n", cmd)
		return nil
	}
	c := exec.Command(cmd[0], cmd[1:]...)
	data, err := c.CombinedOutput()
	os.Stdout.Write(data)
	return err
}

func executeCommandStreaming(cmd []string, stdout, stderr io.Writer) error {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stderr = stderr
	c.Stdout = stdout
	return c.Run()
}
