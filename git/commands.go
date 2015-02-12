package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

//this file contains function do deal with git commands.
// this collection is meant to increase, and eventually get lower level ( use git plumbing API and expose those functions as higher level)

const (
	DefaultTrimCut = "\n \t"
)

//Branch extract the current branch's name (HEAD)
func Branch(prj string) (branch string, err error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), DefaultTrimCut)
	return result, nil
}

//BranchExists returns true if the branch exists
func BranchExists(prj, branch string) (exists bool, err error) {

	refs, err := ForEachRef(prj)
	if err != nil {
		return false, fmt.Errorf("Cannot list git refs. Is this a git repo ? %v", err)
	}

	for _, ref := range refs {
		if ref.Type == CommitType && (ref.Name == branch || ref.Name == "refs/heads/"+branch) {
			// this is the one, it exists
			return true, nil
		}
	}
	return false, nil
}

//Checkout checkout a branch (optionally creating it)
func Checkout(prj, branch string, create bool) (err error) {

	args := make([]string, 0, 3)
	if create {
		args = append(args, "checkout", "-b", branch)
	} else {
		args = append(args, "checkout", branch)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	//log.Printf("%s$ %s\n%s", prj, strings.Join(cmd.Args, " "), out)
	if err != nil {
		return fmt.Errorf("%s: %v", string(out), err)
	}
	return nil
}

//Pull automate the pull with ff only option
func Pull(prj string, ffonly, rebase bool) (result string, err error) {
	args := make([]string, 0, 3)
	args = append(args, "pull")
	if ffonly {
		args = append(args, "--ff-only")
	}
	if rebase {
		args = append(args, "--rebase")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), DefaultTrimCut)
	if err != nil {
		return result, fmt.Errorf("failed to: %s$ git pull : %s", prj, err.Error())
	}
	return result, nil
}

//Clone clone a repo
func Clone(wd, rel, remote, branch string) (result string, err error) {
	cmd := exec.Command("git", "clone", remote, "-b", branch, rel)
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), "\n \t")
	if err != nil {
		return result, fmt.Errorf("failed to git clone %s -b %s %s: %s", remote, branch, rel, err.Error())
	}
	return result, nil
}

//RemoteOrigin returns the current remote.origin.url
// if there is no "origin" remote, then an error is returned.
func RemoteOrigin(prj string) (origin string, err error) {
	return ConfigGet(prj, "remote.origin.url")
}

//RevParseHead read the current commit sha1
func RevParseHead(prj string) (result string, err error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), DefaultTrimCut)
	if err != nil {
		return result, fmt.Errorf("failed to: %s$ git rev-parse HEAD : %s", prj, err.Error())
	}
	return result, nil
}

//RevParseHead read the current commit sha1
func RevListCountHead(prj, branch string) (left, right int, err error) {
	cmd := exec.Command("git", "rev-list", "--count", "--left-right", "HEAD..."+branch)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	result := strings.Trim(string(out), DefaultTrimCut)
	if err != nil {
		return left, right, fmt.Errorf("execution error: %s$ git %s -> error %v: %s", prj, strings.Join(cmd.Args, " "), err, result)
	}
	//log.Printf("git %s : %s", strings.Join(cmd.Args, " "), result)
	split := strings.Split(result, "\t")
	sleft, sright := split[0], split[1]

	left, err = strconv.Atoi(sleft)
	if err != nil {
		return left, right, fmt.Errorf("parsing error: %s$ git %s -> %s: Cannot convert %s to int: %v", prj, strings.Join(cmd.Args, " "), result, sleft, err)
	}
	right, err = strconv.Atoi(sright)
	if err != nil {
		return left, right, fmt.Errorf("parsing error: %s$ git %s -> %s: Cannot convert %s to int: %v", prj, strings.Join(cmd.Args, " "), result, sright, err)
	}
	return left, right, nil
}
