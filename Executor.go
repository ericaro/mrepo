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

type Executor struct {
	wd string //current working dir
	PostProcessor
	DepProcessor
}

func NewExecutor(wd string) *Executor {
	return &Executor{
		wd:            wd,
		PostProcessor: DefaultPostProcessor, //default postprocessor
		DepProcessor:  DepPrinter,           //default depender
	}
}

//Seq run for each `project` in `projects` command `name` with arguments `args`
//
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func (x *Executor) Seq(projects <-chan string, name string, args ...string) {
	var count int
	wd := x.wd
	for prj := range projects {
		count++
		rel, err := filepath.Rel(wd, prj)
		if err != nil {
			log.Fatalf("prj does not appear to be in the current directory %s", err.Error())
		}
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", rel, name, strings.Join(args, " "))
		cmd := exec.Command(name, args...)
		cmd.Dir = prj
		cmd.Stderr, cmd.Stdout, cmd.Stdin = os.Stderr, os.Stdout, os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s %s':\n    %s\n", name, strings.Join(args, " "), err.Error())
		}
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//List just count and print all directories.
func (x *Executor) List(projects <-chan string) {
	wd := x.wd
	var count int
	for prj := range projects {
		count++
		rel, err := filepath.Rel(wd, prj)
		if err != nil {
			rel = prj // uses the absolute path in this case
		}
		fmt.Printf("\033[00;32m%s\033[00m$ \n", rel)
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Concurrent run, in sequences the command on each repository
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func (x *Executor) Concurrent(projects <-chan string, name string, args ...string) {
	wd := x.wd
	outputF := x.PostProcessor

	outputer := make(chan Execution)
	var waiter sync.WaitGroup
	for prj := range projects {
		waiter.Add(1)

		go func(prj string) {
			defer waiter.Done()
			cmd := exec.Command(name, args...)
			cmd.Dir = prj
			out, err := cmd.CombinedOutput()
			if err != nil {
				return
			}
			rel, err := filepath.Rel(wd, prj)
			if err != nil {
				log.Fatalf("prj does not appear to be in the current directory %s", err.Error())
			}
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", prj, name, strings.Join(args, " "))
			//outputer <- head + string(out)
			outputer <- Execution{Name: prj, Rel: rel, Cmd: name, Args: args, Result: string(out)}
		}(prj)
	}

	go func() {
		waiter.Wait()
		close(outputer)
	}()
	outputF(outputer)

}

func (x *Executor) Dependencies(sources <-chan string) {
	wd := x.wd
	dep := x.DepProcessor

	dependencies := make(chan Dependency)
	var waiter sync.WaitGroup

	for prj := range sources {
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
			rel, err := filepath.Rel(wd, prj)
			if err != nil {
				log.Fatalf("prj does not appear to be in the current directory %s", err.Error())
			}
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
