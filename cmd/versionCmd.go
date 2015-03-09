package cmd

import (
	"fmt"
	"os"

	"github.com/ericaro/mrepo/sbr"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(args []string) {

	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(-1, "%v", err)
	}

	v, err := workspace.Version()
	if err != nil {
		fmt.Printf("Cannot compute version: %v\n", err)
	}
	fmt.Printf("%x\n", v)
}
