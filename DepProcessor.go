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

//DependencyProcessor type is called on Dependency to deal with them.
type DependencyProcessor func(prj <-chan Dependency)

//Dependency type contains all the information about each subrepository.
type Dependency struct {
	wd     string
	rel    string //relative path for the project
	remote string
	branch string
}

//DepPrinter simply print out the information, in a tabular way.
func DepPrinter(sources <-chan Dependency) {
	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}

//Cloner actually clone missing dependencies
func Cloner(sources <-chan Dependency, apply bool, w io.Writer) {

	var waiter sync.WaitGroup // to wait for all commands to return
	for d := range sources {
		// if I need to, I will clone

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

//Pruner remove all dependencies locally that are in the chan
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
