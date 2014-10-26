package mrepo

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"
)

//this files contains functions that deals with chan Dependency

//DependencyProcessor type is called on Dependency to deal with them.
type DependencyProcessor func(prj <-chan Dependency)

//Dependency type contains all the information about each subrepository.
type Dependency struct {
	wd     string
	rel    string //relative path for the project
	remote string
	branch string
}

//DependencyPrinter simply print out the information, in a tabular way.
func DependencyPrinter(sources <-chan Dependency) {
	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}

//Cloner clones dependencies
// if apply = false: only a dry run is printed out
// otherwise the operation is made, and printed out.
// results are printed using a tabular format into 'w'
func Cloner(sources <-chan Dependency, apply bool, w io.Writer) {
	var waiter sync.WaitGroup // to wait for all commands to return
	for d := range sources {
		// check if I need to clone
		info, err := os.Stat(filepath.Join(d.wd, d.rel))
		if err == nil && !info.IsDir() {
			// oups there is a file in the way
			log.Fatalf("Cannot clone into %s, a file already exists.")
		}

		if os.IsNotExist(err) { // I need to create one
			waiter.Add(1)
			go func(d Dependency) {
				defer waiter.Done()
				if apply {
					result, err := GitClone(d.wd, d.rel, d.remote, d.branch)
					if err != nil {
						log.Printf("Error during git clone", err.Error())
						return
					}
					fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, result)
				} else {
					fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "DRY RUN")
				}
				//caveat result does not end with a \n so I add it

			}(d)
		}
	}
	waiter.Wait()
}

//Pruner prunes  dependencies
// if apply == false, then only a dry run is printed out.
// otherwise, actually remove the dependency and prints the result
// reults are printed using a tabular format into 'w'
func Pruner(sources <-chan Dependency, apply bool, w io.Writer) {

	var waiter sync.WaitGroup // to wait for all commands to return
	for d := range sources {
		// if I need to, I will clone
		path := filepath.Join(d.wd, d.rel)
		_, err := os.Stat(path)
		if !os.IsNotExist(err) { // it exists
			//schedule a deletion
			waiter.Add(1)
			go func(d Dependency) {
				defer waiter.Done()
				if apply {

					err = os.RemoveAll(filepath.Join(d.wd, d.rel))
					if err != nil {
						log.Printf("Error while pruning tree. %s", err.Error())
						return
					}
					fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "Removing '"+d.rel+"'...")
				} else {
					fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "DRY RUN")
				}
			}(d)
		}
	}
	//wait for all executions
	waiter.Wait()
}

//Diff reads target chan of dependency and current one, an generates two chan
// one for the insertion to be made to current to be equal to target
// one for the deletion to be made to current to be equal to target
//later, maybe we'll add update for branches
func Diff(target, current <-chan Dependency) (insertion, deletion <-chan Dependency) {
	targets := make(map[string]Dependency, 100)
	currents := make(map[string]Dependency, 100)

	ins, del := make(chan Dependency), make(chan Dependency)
	go func() {

		//first flush the targets and currents
		for x := range target {
			targets[x.rel] = x
		}
		for x := range current {
			currents[x.rel] = x
		}

		//then compute the diffs

		for id, t := range targets { // for each target
			_, exists := currents[id]
			if !exists { // if missing , create an insert
				ins <- t
			}
		}
		close(ins)

		for id, c := range currents { // for each current
			_, exists := targets[id]
			if !exists { // locally exists, but not in target, it's a deletion
				del <- c
			}
		}
		close(del)
	}()
	return ins, del
}
