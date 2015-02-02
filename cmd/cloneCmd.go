package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericaro/mrepo"
)

type CloneCmd struct {
	branch *string
}

func (c *CloneCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.branch = fs.String("b", "master", "specify the branch")
	return fs
}

func (c *CloneCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}

	var rel, remote string
	switch len(args) {
	case 0:
		fmt.Printf("Usage sbr clone [-b branch] <remote> [target]\n")
		os.Exit(-1)
	case 1:
		remote = args[0]
		rel = strings.TrimSuffix(filepath.Base(remote), ".git")
	case 2:
		remote = args[0]
		rel = args[1]
	}

	res, err := mrepo.GitClone(wd, rel, remote, *c.branch)
	fmt.Println(res)
	if err != nil {
		fmt.Printf("Error, cannot clone %s: %s\n", remote, err.Error())
		os.Exit(-1)
	}

	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(filepath.Join(wd, rel))
	_, err = workspace.Update(os.Stdout)
	if err != nil {
		fmt.Printf("checkout error: %s", err.Error())
		os.Exit(-1)
	}
}
