package mrepo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//Executor executes the basic commands and dispatch the results to other processors.
//
// It delegates to the Scanner the subrepository lookup, and then:
//
// - executes a random command synchronously or concurrently and pushes the result in a chan of 'Execution'
// asynchronously
//
// - executes builtin git queries about the subrepository and push the result in a chan of 'Dependency'
//
// Caveat, sync execution passes stdin/out to the subprocess that runs the command, so it can run in interactive mode,
// whereas async execution does not.
// async mode is required for statistical postprocessors.
type Executor struct {
	wd string //current working dir
	ExecutionProcessor
	DependencyProcessor
	*scanner
}

//NewExecutor creates a new Executor for a working dir.
func NewExecutor(wd string) *Executor {
	return &Executor{
		wd:                  wd,
		ExecutionProcessor:  DefaultPostProcessor, //default postprocessor
		DependencyProcessor: DepPrinter,           //default depender
		scanner:             newScan(wd),
	}
}

//relpath computes the relative path of a subrepository
func (x *Executor) relpath(subrepository string) string {
	if filepath.IsAbs(subrepository) {
		rel, err := filepath.Rel(x.wd, subrepository)
		if err != nil {
			fmt.Printf("prj does not appear to be in the current directory %s\n", err.Error())
		}
		return rel
	}
	return subrepository
}

//ExecSync runs for each `subrepository` found by Scanner the  command `command` with arguments `args`
// It passes the stdin, stdout, and stderr to the subprocess. and wait for the result.
func (x *Executor) ExecSync(command string, args ...string) {
	var count int
	for sub := range x.Repositories() {
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
func (x *Executor) Exec(command string, args ...string) {
	executions := make(chan Execution)
	var waiter sync.WaitGroup // to wait for all commands to return
	for sub := range x.Repositories() {
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
			executions <- Execution{Name: sub, Rel: rel, Cmd: command, Args: args, Result: string(out)}
		}(sub)
	}

	go func() {
		waiter.Wait()
		close(executions)
	}()
	x.ExecutionProcessor(executions)

}

//Query runs git queries for path, remote url, and branch on each subrepository, and then pushes the result for in a chan of Dependency
func (x *Executor) Query() {
	repositories := x.Repositories()
	wd := x.wd
	dep := x.DependencyProcessor

	dependencies := make(chan Dependency)
	var waiter sync.WaitGroup

	for prj := range repositories {
		waiter.Add(1)
		go func(prj string) {
			defer waiter.Done()

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
					wd:     wd,
					rel:    rel,
					remote: origin,
					branch: branch,
				}
			}
		}(prj)
		//wait and close in a remote so that the main thread ends with the end of processing
	}
	go func() {
		waiter.Wait()
		close(dependencies)
	}()
	dep(dependencies)
}
