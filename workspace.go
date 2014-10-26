package mrepo

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//Workspace represent the current workspace.
//
//
// - executes a random command synchronously
//
// - executes a random command concurrently and pushes the result in a chan of 'Execution'
//
// - read local subrepositories
//
// Caveat, sync execution passes stdin/out to the subprocess that runs the command, so it can run in interactive mode,
// whereas async execution does not.
// async mode is required for statistical postprocessors.
type Workspace struct {
	wd string //current working dir
}

//NewWorkspace creates a new Workspace for a working dir.
func NewWorkspace(wd string) *Workspace {
	return &Workspace{
		wd: wd,
	}
}

//relpath computes the relative path of a subrepository
func (x *Workspace) relpath(subrepository string) string {
	if filepath.IsAbs(subrepository) {
		rel, err := filepath.Rel(x.wd, subrepository)
		if err != nil {
			fmt.Printf("prj does not appear to be in the current directory %s\n", err.Error())
		}
		return rel
	}
	return subrepository
}

//start a gorutine in charge of scanning the local repo AND return the chan that will contain it.
func (x *Workspace) Scan() <-chan string {
	prjc := make(chan string)
	//the subrepo scanner function
	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if f.Name() == ".git" {
				// it's a repository file
				prjc <- filepath.Dir(path)
				//always skip the repository file
				return filepath.SkipDir
			}
		}
		return nil
	}

	go func() {
		defer close(prjc)
		filepath.Walk(x.wd, walker)
	}()
	return prjc
}

//ExecSync runs for each `subrepository` found by Scanner the  command `command` with arguments `args`
// It passes the stdin, stdout, and stderr to the subprocess. and wait for the result.
func (x *Workspace) ExecSync(command string, args ...string) {
	var count int

	for sub := range x.Scan() {
		count++
		rel := x.relpath(sub)
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", rel, command, strings.Join(args, " "))
		cmd := exec.Command(command, args...)
		cmd.Dir = sub
		cmd.Stderr, cmd.Stdout, cmd.Stdin = os.Stderr, os.Stdout, os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s %s':\n    %s\n", command, strings.Join(args, " "), err.Error())
		}
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Exec runs for each `subrepository` found by Scanner the  command `command` with arguments `args`.
// Each command is executed concurrently, and the outputs are collected (both err, and out).
func (x *Workspace) Exec(command string, args ...string) <-chan Execution {
	executions := make(chan Execution)
	var waiter sync.WaitGroup // to wait for all commands to return
	for sub := range x.Scan() {
		waiter.Add(1)

		go func(sub string) {
			defer waiter.Done()
			cmd := exec.Command(command, args...)
			cmd.Dir = sub
			out, err := cmd.CombinedOutput()
			if err != nil {
				return
			}
			rel := x.relpath(sub)
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", sub, command, strings.Join(args, " "))
			//executions <- head + string(out)
			result := string(out)
			result = strings.Trim(result, defaultTrimCut)
			executions <- Execution{Name: sub, Rel: rel, Cmd: command, Args: args, Result: result}
		}(sub)
	}

	go func() {
		waiter.Wait()
		close(executions)
	}()
	return executions
}

//ExecQuery runs git queries for path, remote url, and branch on each subrepository, and then pushes the result for in a chan of Dependency
func (x *Workspace) ExecQuery() <-chan Dependency {
	dependencies := make(chan Dependency)
	go func() {

		for prj := range x.Scan() {

			branch, err := GitBranch(prj)
			if err != nil {
				log.Fatalf("err getting branch %s", err.Error())
			}
			origin, err := GitRemoteOrigin(prj)
			if err != nil {
				log.Fatalf("err getting origin %s", err.Error())
			}
			rel := x.relpath(prj)
			if rel != "." {
				dependencies <- Dependency{
					wd:     x.wd,
					rel:    rel,
					remote: origin,
					branch: branch,
				}
			}
		}
		//wait and close in a remote so that the main thread ends with the end of processing
		close(dependencies)
	}()
	return dependencies
}

//ParseDependencies scans 'r' and fill <-chan Dependency
func (p *Workspace) ParseDependencies(r io.Reader) <-chan Dependency {
	dependencies := make(chan Dependency)
	go func() {

		scanner := bufio.NewScanner(r)
		//use a word splitter
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			rel := scanner.Text()
			if !scanner.Scan() {
				log.Fatalf("missing remote definition.")
			}
			remote := scanner.Text()
			if !scanner.Scan() {
				log.Fatalf("missing branch definition.")
			}
			branch := scanner.Text()

			//log.Printf("scanned to: git clone %s -b %s %s", remote, branch, rel)
			dependencies <- Dependency{
				rel:    rel,
				remote: remote,
				branch: branch,
				wd:     p.wd,
			}

		}
		close(dependencies) //done parsing
	}()
	return dependencies
}
