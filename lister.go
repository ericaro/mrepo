package mrepo

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
)

//Depender type is called on dependency to deal with them.
type Depender func(prj <-chan dependency)

//dependency type contains all the information about each subrepository.
type dependency struct {
	wd     string
	rel    string //relative path for the project
	remote string
	branch string
}

//Deps simply print out the information, in a tabular way.
func Deps(sources <-chan dependency) {
	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}

//Makedeps prints out a Makefile content to rebuild the whole tree.
//There is a goblal `tree` target, that depends on each subrepository.
// And for each subrepository, there is a recipe to build it:
// <path>: ; git clone <remote> -b <branch> $@
func Makedeps(sources <-chan dependency) {
	// print results in two buffers
	// one for  the top recipe, one for each subrepository recipe.
	var topRecipe bytes.Buffer
	var prjRecipe bytes.Buffer
	w := tabwriter.NewWriter(&prjRecipe, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s:\t;git clone\t%q\t-b %q\t$@\n", d.rel, d.remote, d.branch)
		fmt.Fprintf(&topRecipe, "    %s\\\n", d.rel) // mark the prj as a dependency
	}
	w.Flush()
	fmt.Printf("tree: \\\n%s\n%s\n",
		string(topRecipe.Bytes()),
		string(prjRecipe.Bytes()),
	)

}
