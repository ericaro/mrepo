package main

import (
	"github.com/ericaro/help"
	"github.com/ericaro/mrepo/cmd"

	"github.com/rakyll/command"
)

func main() {
	command.On("version",
		"                : compute the sha1 of all dependencies' sha1", &cmd.VersionCmd{}, nil)
	command.On("write",
		"                : write into '.sbr' to reflect directory changes", &cmd.WriteCmd{}, nil)
	command.On("checkout",
		"                : pull top; clone new dependencies; pull all other dependencies (deprecated dependencies can be pruned using -f option)", &cmd.CheckoutCmd{}, nil)
	command.On("compare",
		"                : list directories to be created or deleted", &cmd.CompareCmd{}, nil)
	command.On("merge",
		"                : compare '.sbr' content and directory structure using 'meld'", &cmd.MergeCmd{}, nil)
	command.On("clone",
		"<remote> [path] : clone a remote repository, and then checkout it's .sbr", &cmd.CloneCmd{}, nil)
	command.On("x",
		"<command> <args>: exec arbitrary command on each subrepository", &cmd.ExecCmd{}, nil)
	command.On("status",
		"[revision]      : count commits between HEAD and 'revision'", &cmd.StatusCmd{}, nil)
	command.On("format",
		"                : rewrite current '.sbr' into a cannonical format", &cmd.FormatCmd{}, nil)

	//also declare docs
	command.On("help",
		"[sections...]   : display sections summary, or section details", help.Command, nil)

	help.Section("format", "sbr format description", cmd.SbrFormatMd)

	command.ParseAndRun()
}
