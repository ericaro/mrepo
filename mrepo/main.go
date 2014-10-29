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

  Find subrepositories in both the current directory and the .mrepo file.

  By default, it just list the differences between the two.
	
  Optionally, it is possible to actually clone and prune subrepository to match the definition in the .mrepo dependency file.
  	'-clone' will clone missing dependencies in the working dir.
  	'-prune' will remove extraneous dependencies from the working dir.
  

  With '-update' the differences are presented the other way round, as changed to be applied the the .mrepo file. The meaning of
  the options '-clone' and '-prune' is changed:
  	'-clone' will append items to the current .mrepo file.
  	'-prune' will remove items from .mrepo file.


  With '-list', it only reads an prints the local subrepositories,
  in the .mrepo format. So that, it is possible to just do:

    $ mrepo -list > .mrepo

  Which should be equivalent to:

    $ mrepo -update -clone -prune
  

OPTIONS:

`
	Example = `
EXAMPLES:

  Getting started:

	$ mrepo --list > .mrepo

  You are now ready to share the .mrepo file (within your git workspace for instance). Other developper just need to:

	$ mrepo 

  for a dry run. or 
  
	$ mrepo -clone -prune
  
  for a full apply.

`
)

var (
	list     = flag.Bool("list", false, "print out subrepositories present in the working directory, in the .mrepo format")
	prune    = flag.Bool("prune", false, "actually prune extraneous subrepositories")
	clone    = flag.Bool("clone", false, "actually clone missing subrepositories")
	reverse  = flag.Bool("update", false, "update dependency file, based on information found in the working dir")
	dotmrepo = flag.String("s", ".mrepo", "override default dependency filename")
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
		current := workspace.DependencyWorkingDir()
		// and just print it out
		current.FormatMrepo(os.Stdout)

	case *reverse:
		// the output will be fully tabbed
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		del, ins := workspace.WorkingDirUpdates()
		current := workspace.DependencyFile()
		changed := current.Remove(del, *prune, w)
		changed = current.Add(ins, *clone, w) || changed
		if changed {
			workspace.WriteDependencyFile(current)
		}
		w.Flush()

	default:
		ins, del := workspace.WorkingDirUpdates()
		// the output will be fully tabbed
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		ins.Clone(*clone, w)
		del.Prune(*prune, w)
		w.Flush()
	}
}
