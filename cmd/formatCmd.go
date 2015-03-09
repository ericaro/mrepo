package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/mrepo/sbr"
)

type FormatCmd struct {
	legacy *bool
}

func (c *FormatCmd) Flags(fs *flag.FlagSet) {
	c.legacy = fs.Bool("legacy", false, "format the output using the legacy format")
}

func (c *FormatCmd) Run(args []string) {
	// use wd by default
	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(CodeNoWorkingDir, "%v", err)
	}

	current, err := workspace.Read()

	f, err := os.Create(workspace.Sbrfile())
	if err != nil {
		fmt.Printf("Error Cannot write dependency file: %s", err.Error())
		os.Exit(-1)
	}
	defer f.Close()
	sbr.WriteTo(f, current)
}
