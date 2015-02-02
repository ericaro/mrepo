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
		mrepo.LegacyFmt = true
	}
	WriteSbr(workspace, current)
}

//WriteSbr write down the workspace sbr. Print errors and exit on fail.
// this is not an API!
func WriteSbr(w *mrepo.Workspace, current mrepo.Subrepositories) {
	f, err := os.Create(w.Sbrfile())
	if err != nil {
		fmt.Printf("Error Cannot write dependency file: %s", err.Error())
		os.Exit(-1)
	}
	defer f.Close()
	mrepo.WriteSbr(f, current)
}
