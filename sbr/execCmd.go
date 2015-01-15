package main

import (
	"flag"
	"log"
	"os"

	"github.com/ericaro/mrepo"
)

type execCmd struct {
	cat, sum, count, digest *bool
}

func (c *execCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// output selection
	c.cat = flag.Bool("cat", false, "concatenate outputs, and print it")
	c.sum = flag.Bool("sum", false, "parse each output as a number and print out the total")
	c.count = flag.Bool("count", false, "count different outputs, and prints the resulting histogram")
	c.digest = flag.Bool("digest", false, "compute the sha1 digest of all outputs")

	return fs
}

func (c *execCmd) Run(args []string) {

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//build the workspace, that is used to trigger all commands
	workspace := mrepo.NewWorkspace(wd)

	//again, passing the stdin, and stdout to the subprocess prevent: async, and ability to collect the outputs
	// for special outputers we need to collect outputs, so the 'special' var.
	// special => concurrent mode (because we need to collect outputs)
	// Therefore, selecting the output mode imply selecting "special"= true|false
	// and the ExecutionProcessor function
	xargs := make([]string, 0)
	if len(args) > 1 {
		xargs = args[1:]
	}
	name := args[0]
	executions := workspace.ExecConcurrently(name, xargs...)
	switch {
	case *c.cat:
		mrepo.Cat(executions)
	case *c.sum:
		mrepo.Sum(executions)
	case *c.count:
		mrepo.Count(executions)
	case *c.digest:
		mrepo.Digest(executions)
	default:
		mrepo.ExecutionPrinter(executions)
	}
}
