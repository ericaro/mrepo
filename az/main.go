package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"log"
	"os"
	"path/filepath"
)

// the main that run a command on all sub commands

var async = flag.Bool("a", false, "Controls the execution mode.\n           '-a' or '-a=true' run commands asynchronously.\n           '-a=false' of by default run commands sequentially.")
var list = flag.Bool("l", false, "Dry mode just list the repositories.")

// output selection
var cat = flag.Bool("cat", false, "concatenate outputs, and print it")
var sum = flag.Bool("sum", false, "parse each output as a number and print out the total")
var count = flag.Bool("count", false, "count different outputs, and prints the resulting histogram")
var digest = flag.Bool("digest", false, "compute the sha1 digest of all outputs")

// missing an outputer that takes care of "error codes"

var help = flag.Bool("h", false, "Print this help.")

func main() {
	flag.Parse()
	if (flag.NArg() == 0 && !*list) || *help {
		fmt.Printf(`USAGE %s [-options] <command> <args...>
			
DESCRIPTION:

  Run '<command> <args...>' is every repository found in the current directory hierarchy.

OPTIONS:
	
`, os.Args[0])
		flag.PrintDefaults()

		fmt.Println("\nEXAMPLE:\n")

		fmt.Printf("'%s git status -s'\n", os.Args[0])
		os.Exit(-1)
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error, cannot determine the current directory. %s\n", err.Error())
	}

	executor := mrepo.NewExecutor(wd)

	// parses the remaining args in order to pass them to the underlying process
	args := make([]string, 0)
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}
	name := flag.Arg(0)

	go func() {
		err = executor.Find()
		if err != nil {
			fmt.Printf("Error scanning current directory (%s). %s", wd, err.Error())
		}
	}()

	if *list {
		//for now there is only one way to print dependencies
		//List just count and print all directories.
		var count int
		for prj := range executor.Repositories() {
			count++
			rel, err := filepath.Rel(wd, prj)
			if err != nil {
				rel = prj // uses the absolute path in this case
			}
			fmt.Printf("\033[00;32m%s\033[00m$ \n", rel)
		}
		fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)

	} else {
		// select the output mode

		//again, passing the stdin, and stdout to the subprocess prevent: async, and ability to collect the outputs
		// for special outputers we need to collect outputs, so the 'special' var.
		// special => concurrent mode (because we need to collect outputs)
		// Therefore, selecting the output mode imply selecting "special"= true|false
		// and the ExecutionProcessor function
		var special bool = true

		switch {
		case *cat:
			executor.ExecutionProcessor = mrepo.Cat
		case *sum:
			executor.ExecutionProcessor = mrepo.Sum
		case *count:
			executor.ExecutionProcessor = mrepo.Count
		case *digest:
			executor.ExecutionProcessor = mrepo.Digest
		default:
			special = false
		}

		if special || *async { // this implies concurrent
			// based on the async option, exec asynchronously or sequentially.
			// we cannot just make "seq" a special case of concurrent, since when running sequentially we provide
			// direct access to the std streams. commands can use stdin, and use term escape codes.
			// When in async mode, we just can't do that.
			executor.Exec(name, args...)
		} else {
			executor.ExecSync(name, args...)
		}
	}

}
