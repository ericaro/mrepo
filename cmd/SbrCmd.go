package cmd

import (
	"flag"

	"github.com/ericaro/command"
	"github.com/ericaro/help"
)

type SbrCmd struct {
	command.Commander
}

func (c *SbrCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.Commander = command.NewCommander("sbr", fs)

	c.On("version",
		"                : compute the sha1 of all dependencies' sha1", &VersionCmd{}, nil)
	c.On("write",
		"                : write into '.sbr' to reflect directory changes", &WriteCmd{}, nil)
	c.On("checkout",
		"                : pull top; clone new dependencies; pull all other dependencies (deprecated dependencies can be pruned using -f option)", &CheckoutCmd{}, nil)
	c.On("compare",
		"                : list directories to be created or deleted", &CompareCmd{}, nil)
	c.On("merge",
		"                : compare '.sbr' content and directory structure using 'meld'", &MergeCmd{}, nil)
	c.On("clone",
		"<remote> [path] : clone a remote repository, and then checkout it's .sbr", &CloneCmd{}, nil)
	c.On("x",
		"<command> <args>: exec arbitrary command on each subrepository", &ExecCmd{}, nil)
	c.On("status",
		"[revision]      : count commits between HEAD and 'revision'", &StatusCmd{}, nil)
	c.On("format",
		"                : rewrite current '.sbr' into a cannonical format", &FormatCmd{}, nil)

	// c.On("cilog",
	// 	"                : print remote ci status. Use --tail to tail remote logs", &CilogCmd{}, nil)

	c.On("ci",
		"<command> <args>: remote ci commander. Type 'sbr ci' for help", &CICmd{}, nil)

	//also declare docs
	c.On("help",
		"[sections...]   : display sections summary, or section details", help.Command, nil)

	help.Section("format", "sbr format description", SbrFormatMd)
	return fs
}
