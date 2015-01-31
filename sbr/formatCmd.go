package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo"
)

type formatCmd struct {
}

func (c *formatCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *formatCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	current := workspace.FileSubrepositories()
	workspace.WriteSubrepositoryFile(current)

	fmt.Printf("Done")
}
