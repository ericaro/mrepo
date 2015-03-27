package cmd

import (
	"fmt"
	"os"

	"github.com/ericaro/sbr/sbr"
)

type FetchCmd struct {
}

func (c *FetchCmd) Run(args []string) {

	//get the revision to compare to (defaulted to origin/master)

	//creates a workspace to be able to read from/to sets
	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(-1, "%v", err)
	}
	fmt.Printf("Fetching all...")

	executions := ExecConcurrently(workspace, "git", "fetch")
	ExecutionPrinter(executions)
}
