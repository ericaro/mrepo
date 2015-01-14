package main

import (
	"github.com/rakyll/command"
)

func main() {
	command.On("version", "compute the sha1 of all dependencies' sha1", &versionCmd{}, nil)
	command.On("write", "write into '.sbr' to reflect directory changes", &writeCmd{}, nil)
	command.On("checkout", "pull top; clone new dependencies; pull all other dependencies (deprecated dependencie can be pruned using -f option)", &checkoutCmd{}, nil)
	command.On("compare", "list directories to be created or deleted", &compareCmd{}, nil)
	command.On("merge", "compare '.sbr' content and directory structure using 'meld'", &mergeCmd{}, nil)
	command.On("clone", "clone a remote repository, and then checkout it's .sbr", &cloneCmd{}, nil)
	command.ParseAndRun()
}
