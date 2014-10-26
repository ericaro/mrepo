package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"os"
)

var (
	makefile = flag.Bool("makefile", false, "Print dependencies in a Makefile format")
	help     = flag.Bool("h", false, "Print this help.")
)

func main() {
	flag.Parse()
	if *help {
		fmt.Println(`USAGE git deps [-options]
			
DESCRIPTION:

  Manage git dependencies.
  Scan recursively the current directory, looking for embedded git repositories.
  By default, it just prints out each one in a tabular way.

  You can print them in a Makefile format (-makefile)


OPTIONS:
`)
		flag.PrintDefaults()

		fmt.Println(`
EXAMPLES:

git deps

git deps -makefile > Makefile
`)
		os.Exit(-1)
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
	}
	executor := mrepo.NewExecutor(wd)
	//scan for current wd
	go func() {
		err = executor.Find()
		if err != nil {
			fmt.Printf("Error scanning current directory (%s). %s", wd, err.Error())
		}
	}()

	//select the "Depender" i.e what to do with each repo
	switch {
	case *makefile:
		executor.DependencyProcessor = mrepo.Makefiler
	default:
		executor.DependencyProcessor = mrepo.DepPrinter
	}
	executor.Query()
}
