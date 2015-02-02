package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo"
)

type CheckoutCmd struct {
	force *bool
}

func (c *CheckoutCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.force = fs.Bool("f", false, "force prune")
	return fs
}

func (c *CheckoutCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}

	if *c.force {
		fmt.Printf("PRUNE mode\n")
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)
	err = workspace.PullTop(os.Stdout)
	if err != nil {
		fmt.Printf("Failed to pull top: %s", err.Error())
		os.Exit(-1)
	}
	if *c.force {
		_, err = workspace.Refresh(os.Stdout)
	} else {
		_, err = workspace.Update(os.Stdout)
	}
	if err != nil {
		fmt.Printf("checkout error: %s", err.Error())
		os.Exit(-1)
	}
}
