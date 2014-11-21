package mrepo

import (
	"fmt"
	"io"
	"os/exec"
)

//Make invoke make on the prj with the target as argument.
func Make(prj, target string, buf io.Writer) (err error) {

	//because, I don't know why, but $(PWD) will return "os.Getwd" instead of prj
	// I need to go through a bash
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && make %s", prj, target))
	cmd.Dir = prj
	cmd.Stdout = buf
	cmd.Stderr = buf
	return cmd.Run()
}

//Meld invoke meld on the prj with the target as argument.
func Meld(wd, title, file1, file2 string) (err error) {
	cmd := exec.Command("meld", "-L", title, file1, file2)
	cmd.Dir = wd
	return cmd.Run()
}
