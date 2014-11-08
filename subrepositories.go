package mrepo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"
)

//this files contains functions that deals with subrepositories

//Subrepository type contains all the information about a subrepository.
type Subrepository struct {
	wd     string // absolute path for the working dir
	rel    string //relative path for the project
	remote string
	branch string
}

//Rel returns this project's relative path.
func (d *Subrepository) Rel() string {
	return d.rel
}

//Subrepositories represent a set of subrepositories.
// Subrepositories are always stored sorted by "rel"
type Subrepositories []Subrepository

func (a Subrepositories) Len() int           { return len(a) }
func (a Subrepositories) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Subrepositories) Less(i, j int) bool { return a[i].rel < a[j].rel }

//DependencyPrinter simply print out the information, in a tabular way.
func (d *Subrepositories) Print(out io.Writer) {
	sources := *d
	w := tabwriter.NewWriter(out, 3, 8, 3, ' ', 0)
	for _, d := range sources {
		fmt.Fprintf(w, "git\t%q\t%q\t%q\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}

//Add append a bunch of subrepositories to 'd'
func (d *Subrepositories) Add(ins Subrepositories, apply bool, w io.Writer) (changed bool) {
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

//Remove subrepositories from 'd'
// apply make this method act like a dry run
func (d *Subrepositories) Remove(del Subrepositories, apply bool, w io.Writer) (changed bool) {
	sources := *d
	deleted := indexSbr(del)
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

//Clone subrepositories into the working directory.
// if apply = false: only a dry run is printed out
// otherwise the operation is made, and printed out.
// results are printed using a tabular format into 'w'
func (d *Subrepositories) Clone(apply bool, w io.Writer) error {
	sources := *d
	var waiter sync.WaitGroup // to wait for all commands to return
	var cloneerror error
	for _, d := range sources {
		// check if I need to clone
		info, err := os.Stat(filepath.Join(d.wd, d.rel))
		if err == nil && !info.IsDir() {
			// oups there is a file in the way
			return fmt.Errorf("Cannot clone into %s, a file already exists.")
		}
		if os.IsNotExist(err) { // I need to create one
			waiter.Add(1)
			go func(d Subrepository) {
				defer waiter.Done()
				if apply {
					result, err := GitClone(d.wd, d.rel, d.remote, d.branch)
					if err != nil {
						if cloneerror != nil { // keep the first error
							cloneerror = err
						}
						fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, result, err.Error())
					} else {
						fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, result)
					}
				} else {
					fmt.Fprintf(w, "clone\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "DRY RUN")
				}
				//caveat result does not end with a \n so I add it

			}(d)
		}
	}
	waiter.Wait()
	return cloneerror
}

//Prune dependencies from the working directory.
// if apply == false, then only a dry run is printed out.
// otherwise, actually remove the dependency and prints the result
// reults are printed using a tabular format into 'w'
func (d *Subrepositories) Prune(apply bool, w io.Writer) error {
	sources := *d
	var waiter sync.WaitGroup // to wait for all commands to return
	var pruneerror error
	for _, d := range sources {
		// if I need to, I will clone
		path := filepath.Join(d.wd, d.rel)
		_, err := os.Stat(path)
		if !os.IsNotExist(err) { // it exists
			//schedule a deletion
			waiter.Add(1)
			go func(d Subrepository) {
				defer waiter.Done()
				if apply {

					err = os.RemoveAll(filepath.Join(d.wd, d.rel))
					if err != nil {
						fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "Removing '"+d.rel+"'...", err.Error())
					} else {
						fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "Removing '"+d.rel+"'...")
					}
				} else {
					fmt.Fprintf(w, "prune\t%s\t%s\t%s\t%s\n", d.rel, d.remote, d.branch, "DRY RUN")
				}
			}(d)
		}
	}
	//wait for all executions
	waiter.Wait()
	return pruneerror
}

//Diff compute the changes to be applied to 'current', in order to became target.
// updates are not handled, just insertion, and deletion.
//later, maybe we'll add update for branches
func (current Subrepositories) Diff(target Subrepositories) (insertion, deletion Subrepositories) {
	ins, del := make([]Subrepository, 0, 100), make([]Subrepository, 0, 100)
	targets := indexSbr(target)
	currents := indexSbr(current)

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

//indexSbr build up a small index of Subrepository based on their .rel attribute.
func indexSbr(deps []Subrepository) map[string]Subrepository {
	i := make(map[string]Subrepository, 100)
	for _, x := range deps {
		i[x.rel] = x
	}
	return i
}
