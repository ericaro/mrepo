package mrepo

import (
	"io"
	"os/exec"
)

//Make invoke make on the prj with the target as argument.
func Make(prj, target string, buf io.Writer) (err error) {
	cmd := exec.Command("make", target)
	cmd.Dir = prj
	cmd.Stdout = buf
	cmd.Stderr = buf
	return cmd.Run()
}
