package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo"
)

type CheckoutCmd struct {
	prune  *bool
	ffonly *bool
	rebase *bool
}

func (c *CheckoutCmd) Flags(fs *flag.FlagSet) {
	c.prune = fs.Bool("prune", false, "prune sub repositories that are not in the .sbr file")
	c.ffonly = fs.Bool("ff-only", false, "Refuse to merge and exit with a non-zero status unless the current HEAD is already up-to-date or the merge can be resolved as a fast-forward.")
	c.rebase = fs.Bool("rebase", false, "rebase instead of merge")
}

func (c *CheckoutCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()

	if *c.prune {
		fmt.Printf("PRUNE mode\n")
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)
	_, err := workspace.Checkout(os.Stdout, *c.prune, *c.ffonly, *c.rebase)
	if err != nil {
		fmt.Printf("checkout error: %s", err.Error())
		os.Exit(-1)
	}
}
