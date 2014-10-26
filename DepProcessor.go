package mrepo

import (
	"bytes"
	"fmt"
	"os"
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

// TODO(EA) add a Maker that will actually run the git clone.
// the trick is tha the "scanner" will need to read from a source file rather than the disk.
