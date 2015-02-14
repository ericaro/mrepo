package cmd

import (
	"flag"

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
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	current := workspace.FileSubrepositories()
	if *c.legacy {
		mrepo.LegacyFmt = true
	}
	WriteSbr(workspace, current)
}
