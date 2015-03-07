package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ericaro/mrepo"
)

type MergeCmd struct {
}

func (c *MergeCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	//generate a temp file
	current := workspace.WorkingDirSubrepositories()
	f, err := ioutil.TempFile("", "sbr")
	if err != nil {
		fmt.Printf("Cannot generate temp file: %s", err.Error())
		os.Exit(-1)

	}
	mrepo.WriteSbr(f, current)
	f.Close() //no defer to open it up just after.
	err = mrepo.Meld(workspace.Wd(), ".sbr set  |  disk set", workspace.Sbrfile(), f.Name())
	if err != nil {
		fmt.Printf("Meld returned with error: %s", err.Error())
		os.Exit(-1)
	}
	// shall I apply ?
}
