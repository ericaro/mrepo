package main

import (
	"github.com/ericaro/help"

	"github.com/rakyll/command"
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
	command.On("format",
		"                : rewrite current '.sbr' into a cannonical format", &formatCmd{}, nil)

	//also declare docs
	command.On("help",
		"[sections...]   : display sections summary, or section details", help.Command, nil)

	help.Section("format", "sbr format description", sbrFormatMd)

	command.ParseAndRun()
}
