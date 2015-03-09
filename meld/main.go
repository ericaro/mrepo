package meld

import "os/exec"

//Meld invoke meld on the prj with the target as argument.
func Diff(wd, title, file1, file2 string) (err error) {
	cmd := exec.Command("meld", "-L", title, file1, file2)
	cmd.Dir = wd
	return cmd.Run()
}
