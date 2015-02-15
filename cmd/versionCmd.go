package cmd

import (
	"flag"
	"fmt"

	"github.com/ericaro/mrepo"
)

type VersionCmd struct{}

func (c *VersionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *VersionCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)
	v, err := workspace.Version()
	if err != nil {
		fmt.Printf("Cannot compute version: %v\n", err)
	}
	fmt.Printf("%x\n", v)
}

//byName to sort any slice of Execution by their Name !
type byName []string

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i] < a[j] }
