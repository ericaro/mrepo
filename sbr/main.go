package main

import "github.com/rakyll/command"

const (
	Usage = `USAGE sbr <command> [options] [args]

'sbr' is a workspace subrepository manager

It helps you deal with a workspace made of several 'git' repositories.


`
)

func main() {
	command.On("version",
		"                : compute the sha1 of all dependencies' sha1", &versionCmd{}, nil)
	command.On("write",
		"                : write into '.sbr' to reflect directory changes", &writeCmd{}, nil)
	command.On("checkout",
		"                : pull top; clone new dependencies; pull all other dependencies (deprecated dependencies can be pruned using -f option)", &checkoutCmd{}, nil)
	command.On("compare",
		"                : list directories to be created or deleted", &compareCmd{}, nil)
	command.On("merge",
		"                : compare '.sbr' content and directory structure using 'meld'", &mergeCmd{}, nil)
	command.On("clone",
		"<remote> [path] : clone a remote repository, and then checkout it's .sbr", &cloneCmd{}, nil)
	command.On("x",
		"<command> <args>: exec arbitrary command on each subrepository", &execCmd{}, nil)
	command.On("status",
		"[revision]      : count commits between HEAD and 'revision'", &statusCmd{}, nil)

	command.ParseAndRun()
}
