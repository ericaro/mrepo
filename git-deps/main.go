package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"os"
	"sync"
	"text/tabwriter"
)

var (
	prune = flag.Bool("prune", false, "In diff mode, actually prune extraneous subrepositories")
	clone = flag.Bool("clone", false, "In diff mode, actually clone missing subrepositories")
	diff  = flag.Bool("diff", false, "activate diff mode. Compare working dir subrepositories, and the one read in stdin.")
	help  = flag.Bool("h", false, "Print this help.")
)

func usage() {

	fmt.Println(`USAGE git deps [-options]
			
DESCRIPTION:

  Manage git dependencies.

  Scan recursively the current directory, looking for embedded git repositories.
  By default, it just prints out each one in a tabular format.

  In diff mode, it also reads dependencies from stdin, in the same tabular format.

  It then compare local subrepositories, and target subrepositories, to build a list
  of insertion/deletion.

  By default, it just prints out this changes.

  Using '-clone' you can actually clone insertions.
  Using '-prune' you can actually prune deletions.


OPTIONS:
`)
	flag.PrintDefaults()

	fmt.Println(`
EXAMPLES:
	

In one workspace, read local subrepositories.
    $ git deps > subrepos

In another workspace, compare with previous
    $ git deps -diff < subrepos

Apply changes:
    $ git deps -diff -clone -prune < subrepos

`)
}

func main() {
	flag.Parse()

	if *help {
		usage()
		return
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
	}
	workspace := mrepo.NewWorkspace(wd)

	if *diff {
		//get the chan of dependencies as read from the stdin
		target := workspace.ParseDependencies(os.Stdin) // for now, just parse

		current := workspace.ExecQuery()
		//convert target / current  into insertion, deletion
		ins, del := mrepo.Diff(target, current)

		// the output will be fully tabbed
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		var waiter sync.WaitGroup
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			mrepo.Cloner(ins, *clone, w)
		}()
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			mrepo.Pruner(del, *prune, w)
		}()
		waiter.Wait()
		w.Flush()

	} else { // not diff mode, hence, plain local mode
		// execute query on each subrepo
		current := workspace.ExecQuery()
		// and just print it out
		mrepo.DependencyPrinter(current)
		return
	}
}
