package mrepo

import (
	"bytes"
	"fmt"
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

//Makefiler prints out a Makefile content to rebuild the whole tree.
//There is a goblal `tree` target, that depends on each subrepository.
// And for each subrepository, there is a recipe to build it:
// <path>: ; git clone <remote> -b <branch> $@
func Makefiler(sources <-chan Dependency) {
	// print results in two buffers
	// one for  the top recipe, one for each subrepository recipe.
	var topRecipe bytes.Buffer
	var prjRecipe bytes.Buffer
	w := tabwriter.NewWriter(&prjRecipe, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s:\t;git clone\t%q\t-b %q\t$@\n", d.rel, d.remote, d.branch)
		fmt.Fprintf(&topRecipe, "    %s\\\n", d.rel) // mark the prj as a Dependency
	}
	w.Flush()
	fmt.Printf("tree: \\\n%s\n%s\n",
		string(topRecipe.Bytes()),
		string(prjRecipe.Bytes()),
	)

}

//Cloner actually clone missing dependencies
func Cloner(sources <-chan Dependency) {

	executions := make(chan Execution)
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
				result, err := GitClone(d.wd, d.rel, d.remote, d.branch)
				if err != nil {
					log.Printf("Error in Git Clone", err.Error())
					return
				}
				//caveat result does not end with a \n so I add it
				result = result + "\n"
				executions <- Execution{Name: filepath.Join(d.wd, d.rel), Rel: "", Cmd: "git", Args: []string{"clone", d.remote, "-b", d.branch, d.rel}, Result: result}
			}(d)
		}
	}
	//wait for all executions
	go func() {
		waiter.Wait()
		close(executions)
	}()
	//simply print out
	DefaultPostProcessor(executions)

	//log.Printf("scanned to: git clone %s -b %s %s", remote, branch, rel)

}
