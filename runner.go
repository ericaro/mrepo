package mrepo

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
)

//Seq run, in sequences the command on each project
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func Seq(projects <-chan string, name string, args ...string) {
	var count int
	for prj := range projects {
		count++
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", prj, name, strings.Join(args, " "))
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
func List(projects <-chan string) {
	var count int
	for prj := range projects {
		count++
		fmt.Printf("\033[00;32m%s\033[00m$ \n", prj)
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Replay generate an output able to "replay" the current structure.
// it's a makefile representing the current directory tree, and the way to rebuild it.
// for instance
// tree: dir1 dir2
// dir1:   ; git clone git@github.com/src1 -b prod $@
// dir2:   ; git clone git@github.com/src2 -b dev  $@
func Replay(projects <-chan string, wd string) {
	var topRule bytes.Buffer
	var prjRule bytes.Buffer
	w := tabwriter.NewWriter(&prjRule, 3, 8, 3, ' ', 0)

	for prj := range projects {
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
			fmt.Fprintf(os.Stderr, "prj does not appear to be in the current directory: %s %s", wd, prj)
		} else if rel != "." {
			fmt.Fprintf(w, "%s:\t;git clone\t%q\t-b %q\t$@\n", rel, origin, branch)
			fmt.Fprintf(&topRule, "%s ", rel) // mark the prj as a dependency
		}

	}
	w.Flush()
	fmt.Printf(`
tree: %s
%s`,
		string(topRule.Bytes()),
		string(prjRule.Bytes()),
	)

}

//Concurrent run, in sequences the command on each repository
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func Concurrent(projects <-chan string, shouldPrint bool, outputF PostProcessor, name string, args ...string) {

	var slot string // a reserved space to print and delete messages
	if shouldPrint {
		slot = strings.Repeat(" ", 80)
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", "<for all>", name, strings.Join(args, " "))
	}

	outputer := make(chan execution)
	var waiter sync.WaitGroup
	for prj := range projects {
		waiter.Add(1)

		if shouldPrint {
			fmt.Print("\r    start ")
			if len(prj) > len(slot) {
				fmt.Printf("%s ...", prj[0:len(slot)])
			} else {
				fmt.Printf("%s ...%s", prj, slot[len(prj):])
			}
		}

		go func(prj string) {
			defer waiter.Done()
			cmd := exec.Command(name, args...)
			cmd.Dir = prj
			out, err := cmd.CombinedOutput()
			if err != nil {
				return
			}
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", prj, name, strings.Join(args, " "))
			//outputer <- head + string(out)
			outputer <- execution{Name: prj, Cmd: name, Args: args, Result: string(out)}
		}(prj)
	}
	if shouldPrint {
		fmt.Printf("\r    all started. waiting for tasks to complete...%s\n\n", slot)
	}

	go func() {
		waiter.Wait()
		close(outputer)
	}()
	outputF(outputer)

}
