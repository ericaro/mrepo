package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"os"
	"text/tabwriter"
)

const (
	Usage = `
USAGE %s [-options]
			
DESCRIPTION:

  Manage git dependencies.

  Find subrepositories in both the current directory and the .sbr file.

  By default, it just list the differences between the two.
	
  Optionally, it is possible to actually clone and prune subrepository to match the definition in the .sbr dependency file.
  	'-clone' will clone missing dependencies in the working dir.
  	'-prune' will remove extraneous dependencies from the working dir.
  

  With '-update' the differences are presented the other way round, as changed to be applied the the .sbr file. The meaning of
  the options '-clone' and '-prune' is changed:
  	'-clone' will append items to the current .sbr file.
  	'-prune' will remove items from .sbr file.


  With '-list', it only reads an prints the local subrepositories,
  in the .sbr format. So that, it is possible to just do:

    $ sbr -list > .sbr

  Which should be equivalent to:

    $ sbr -update -clone -prune
  

OPTIONS:

`
	Example = `
EXAMPLES:

  - Init a workspace:

	$ sbr -init

  Fills the .sbr file with subrepositories found on the working dir.

  - Add a dependency, via the .sbr file:
  
	$ echo ' git "src/ericaro/mrepo" "git@github.com:ericaro/mrepo.git" "dev"' >> .sbr
	$ sbr -clone
  
  Clone in the working directory what need to be cloned. 
  Note that the change in the .sbr file could be come from other developpers, via a git pull.

  Add a subrepository, and update your .sbr:

  	$ git clone git@github.com:ericaro/mrepo.git -b dev src/ericaro/mrepo
	$ sbr -update -clone
  
  Now, your .sbr file contains the new dependency. Commit & Push it so teammate will be able to clone it too.



`
)

var (
	list     = flag.Bool("list", false, "print out subrepositories present in the working directory, in the .sbr format")
	prune    = flag.Bool("prune", false, "actually prune extraneous subrepositories")
	clone    = flag.Bool("clone", false, "actually clone missing subrepositories")
	update   = flag.Bool("update", false, "update dependency file, based on information found in the working dir")
	initf    = flag.Bool("init", false, "alias for -update -clone, on an empty directory will just create the .sbr file.")
	dotmrepo = flag.String("s", ".sbr", "override default dependency filename")
	// workingdir = flag.String("wd", ".", "path to be used as working dir")
	help = flag.Bool("h", false, "Print this help.")
)

func usage() {
	fmt.Printf(Usage, os.Args[0])
	flag.PrintDefaults()
	fmt.Printf(Example, os.Args[0])
}

func main() {
	flag.Parse()
	if *initf {
		*update, *clone = true, true
	}

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

	switch {
	case *list: // not diff mode, hence, plain local mode
		// execute query on each subrepo
		current := workspace.WorkingDirSubrepositories()
		// and just print it out
		current.Print(os.Stdout)

	case *update:
		// the output will be fully tabbed
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		del, ins := workspace.WorkingDirPatches()
		current := workspace.FileSubrepositories()
		changed := current.Remove(del, *prune, w)
		changed = current.Add(ins, *clone, w) || changed
		if changed {
			workspace.WriteSubrepositoryFile(current)
		}
		w.Flush()

	default:
		ins, del := workspace.WorkingDirPatches()
		// the output will be fully tabbed
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		ins.Clone(*clone, w)
		del.Prune(*prune, w)
		w.Flush()
	}
}
