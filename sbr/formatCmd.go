package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo"
)

type formatCmd struct {
	legacy *bool
}

func (c *formatCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.legacy = fs.Bool("legacy", false, "format the output using the legacy format")
	return fs
}

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
	workspace.WriteSubrepositoryFileLegacy(current)

	fmt.Printf("Done")
}
