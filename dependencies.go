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

func (d *Dependency) Rel() string {
	return d.rel
}

//Dependencies represent a set of dependencies. Dependencies are always stored in path order.
type Dependencies []Dependency

func (a Dependencies) Len() int           { return len(a) }
func (a Dependencies) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Dependencies) Less(i, j int) bool { return a[i].rel < a[j].rel }

//DependencyPrinter simply print out the information, in a tabular way.
func (d *Dependencies) FormatMrepo(out io.Writer) {
	sources := *d
	w := tabwriter.NewWriter(out, 3, 8, 3, ' ', 0)
	for _, d := range sources {
		fmt.Fprintf(w, "git\t%q\t%q\t%q\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}

//Add append a bunch of dependencies to 'd'
func (d *Dependencies) Add(ins Dependencies, apply bool, w io.Writer) (changed bool) {
	sources := *d
	res := "Dry Run"
	if apply {
		res = "Applied"
	}
	for _, d := range ins {
		if apply {
			sources = append(sources, d)
			changed = true
		}
		fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, res)
	}
	*d = sources
	return
}

//Remove Dependencies from 'd'
// apply make this method act like a dry run
func (d *Dependencies) Remove(del Dependencies, apply bool, w io.Writer) (changed bool) {
	sources := *d
	deleted := dependencyIndex(del)
	result := "Dry Run"
	if apply {
		result = "Applied"
	}

	j := 0
	for i, d := range sources {
		if _, del := deleted[d.rel]; !del { // we simply copy the values, deletion is just an offset in fact
			if i != j && apply { // if apply is false, then sources will never be changed
				sources[j] = sources[i]
				changed = true
			}
			j++
		} else { // to be deleted
			fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, result)
		}
	}
	*d = sources[0:j]
	return
}

//Clone dependencies into the working directory.
// if apply = false: only a dry run is printed out
// otherwise the operation is made, and printed out.
// results are printed using a tabular format into 'w'
func (d *Dependencies) Clone(apply bool, w io.Writer) {
	sources := *d
	var waiter sync.WaitGroup // to wait for all commands to return
	for _, d := range sources {
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

//Prune  dependencies from the working directory.
// if apply == false, then only a dry run is printed out.
// otherwise, actually remove the dependency and prints the result
// reults are printed using a tabular format into 'w'
func (d *Dependencies) Prune(apply bool, w io.Writer) {
	sources := *d
	var waiter sync.WaitGroup // to wait for all commands to return
	for _, d := range sources {
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

//Diff compute the changes to be applied to 'current', in order to became target.
// updates are not handled, just insertion, and deletion.
//later, maybe we'll add update for branches
func (current Dependencies) Diff(target Dependencies) (insertion, deletion Dependencies) {
	ins, del := make([]Dependency, 0, 100), make([]Dependency, 0, 100)
	targets := dependencyIndex(target)
	currents := dependencyIndex(current)

	//then compute the diffs
	for id, t := range targets { // for each target
		_, exists := currents[id]
		if !exists { // if missing , create an insert
			ins = append(ins, t)
		}
	}
	for id, c := range currents { // for each current
		_, exists := targets[id]
		if !exists { // locally exists, but not in target, it's a deletion
			del = append(del, c)
		}
	}
	return ins, del
}

//dependencyIndex build up a small index of Dependency based on their .rel attribute.
func dependencyIndex(deps []Dependency) map[string]Dependency {
	i := make(map[string]Dependency, 100)
	for _, x := range deps {
		i[x.rel] = x
	}
	return i
}
