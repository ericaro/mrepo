package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo"
)

type FormatCmd struct {
	legacy *bool
}

func (c *FormatCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.legacy = fs.Bool("legacy", false, "format the output using the legacy format")
	return fs
}

func (c *FormatCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	current := workspace.FileSubrepositories()
	if *c.legacy {
		workspace.WriteSubrepositoryFileLegacy(current)
	} else {
		workspace.WriteSubrepositoryFile(current)
	}

	fmt.Printf("Done")
}
