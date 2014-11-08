package mrepo

import (
	"os/exec"
	"strings"
)

//Make invoke make on the prj with the target as argument.
func Make(prj, target string) (result string, err error) {
	cmd := exec.Command("make", target)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result = strings.Trim(string(out), "\n \t")
	return result, nil
}
