package mrepo

import (
	"os/exec"
	"strings"
)

//GitBranch extract the current branch name (HEAD)
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

//GitRemoteOrigin returns the current remote.origin.url
// if there is no "origin" remote, then an error is returned.
func GitRemoteOrigin(prj string) (origin string, err error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), "\n \t")
	return result, nil
}