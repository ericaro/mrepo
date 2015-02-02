package cmd

import (
	"flag"
	"log"
	"os"

	"github.com/ericaro/mrepo"
)

type ExecCmd struct {
	cat, sum, count, digest *bool
}

func (c *ExecCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// output selection
	c.cat = fs.Bool("cat", false, "concatenate outputs, and print it")
	c.sum = fs.Bool("sum", false, "parse each output as a number and print out the total")
	c.count = fs.Bool("count", false, "count different outputs, and prints the resulting histogram")
	c.digest = fs.Bool("digest", false, "compute the sha1 digest of all outputs")

	return fs
}

func (c *ExecCmd) Run(args []string) {

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
		mrepo.ExecutionCat(executions)
	case *c.sum:
		mrepo.ExecutionSum(executions)
	case *c.count:
		mrepo.ExecutionCount(executions)
	case *c.digest:
		mrepo.ExecutionDigest(executions)
	default:
		mrepo.ExecutionPrinter(executions)
	}
}
