package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"os"
	"sync"
	"text/tabwriter"
)

// by defautl mrepo : prints out a diff
// with --prune or --clone applies deletions and insertions
// I need to be a bit smarter while reading the file:
// skip comments (#)
// add nature (git) in order to open up to other kind of repo
// make sure that the list
var (
	list     = flag.Bool("list", false, "print out subrepositories present in the working directory, in a format suitable for .mrepo file")
	prune    = flag.Bool("prune", false, "actually prune extraneous subrepositories")
	clone    = flag.Bool("clone", false, "actually clone missing subrepositories")
	dotmrepo = flag.String("-s", ".mrepo", "replacement for .mrepo filename")
	// workingdir = flag.String("wd", ".", "path to be used as working dir")
	help = flag.Bool("h", false, "Print this help.")
)

func usage() {

	fmt.Printf(`USAGE %s [-options]
			
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

`, os.Args[0])
	flag.PrintDefaults()

	fmt.Println(`
EXAMPLES:

  Getting started:

	$ mrepo --list > .mrepo

  You are now ready to share the .mrepo file (within your git workspace for instance). Other developper just need to:

	$ mrepo 

  for a dry run.

  Add a new subrepository. 
  Use your own git clone, instead of the original repository, albeit in the same place it would have beeen.

    $ echo 'git "src/github.com/ericaro/mrepo" "git@github.com:myself/mrepo.git"  "dev"' >> .mrepo
    $ mrepo --clone

  And all other developpers just need to do:

	$ git pull
	$ mrepo --clone

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

	if *list { // not diff mode, hence, plain local mode
		// execute query on each subrepo
		current := workspace.ExecQuery()
		// and just print it out
		mrepo.MrepoFormat(current)
	} else {

		//get the chan of dependencies as read from .mrepo

		var target <-chan mrepo.Dependency
		file, err := os.Open(*dotmrepo)
		if err != nil {
			fmt.Printf("Cannot read .mrepo file %s", *dotmrepo, err.Error())
			// just print out the error, and init an empty .mrepo file
		} else {
			defer file.Close()
			target = workspace.ParseDependencies(file) // for now, just parse
		}

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
	}

}
