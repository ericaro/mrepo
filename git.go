package mrepo

import (
	"os/exec"
	"strings"
)

func GitBranch(prj string) (branch string, err error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), "\n \t")
	return result, nil
}

func GitRemoteOrigin(prj string) (branch string, err error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), "\n \t")
	return result, nil
}
