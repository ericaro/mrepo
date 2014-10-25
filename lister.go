package mrepo

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
)

type dependency struct {
	wd     string
	rel    string //relative path for the project
	remote string
	branch string
}

//Depender is used to deal with a bunch of dependencies
type Depender func(prj <-chan dependency)

func Deps(sources <-chan dependency) {
	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.rel, d.remote, d.branch)
	}
	w.Flush()
}
func Makedeps(sources <-chan dependency) {
	// print results in two buffers // one for single prject rule, one for the top dependency
	var topRule bytes.Buffer
	var prjRule bytes.Buffer
	w := tabwriter.NewWriter(&prjRule, 3, 8, 3, ' ', 0)

	for d := range sources {
		fmt.Fprintf(w, "%s:\t;git clone\t%q\t-b %q\t$@\n", d.rel, d.remote, d.branch)
		fmt.Fprintf(&topRule, "    %s\\\n", d.rel) // mark the prj as a dependency
	}
	w.Flush()
	fmt.Printf("tree: \\\n%s\n%s\n",
		string(topRule.Bytes()),
		string(prjRule.Bytes()),
	)

}
