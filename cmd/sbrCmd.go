package cmd

import (
	"flag"

	"github.com/ericaro/mrepo/format"

	"github.com/ericaro/command"
	"github.com/ericaro/help"
)

type SbrCmd struct {
	command.Commander
}

func (c *SbrCmd) Flags(fs *flag.FlagSet) {
	c.On("version", "", "compute the sha1 of all dependencies' sha1", &VersionCmd{})
	c.On("write", "", "write into '.sbr' to reflect directory changes", &WriteCmd{})
	c.On("checkout", "", "pull top; clone new dependencies; pull all other dependencies (deprecated dependencies can be pruned using -f option)", &CheckoutCmd{})
	c.On("compare", "", "list directories to be created or deleted", &CompareCmd{})
	c.On("merge", "", "compare '.sbr' content and directory structure using 'meld'", &MergeCmd{})
	c.On("clone", "<remote> [path]", "clone a remote repository, and then checkout it's .sbr", &CloneCmd{})
	c.On("x", "<command> <args>", "exec arbitrary command on each subrepository", &ExecCmd{})
	c.On("status", "[revision]", "count commits between HEAD and 'revision'", &StatusCmd{})
	c.On("format", " ", "rewrite current '.sbr' into a cannonical format", &FormatCmd{})

	// c.On("cilog",
	// 	" "                ,"print remote ci status. Use --tail to tail remote logs", &CilogCmd{}, nil)

	c.On("ci", "<command> <args>", "remote ci commander. Type 'sbr ci' for help", &CICmd{command.NewCommander()})

	//also declare docs
	c.On("help", "[sections...]", "display sections summary, or section details", help.Command)

	help.Section("format", "sbr format description", SbrFormatMd)
	help.Section("ci", "CI server manual", CIServerMd)
	help.Section("protocol", "CI server protocol", string(format.CIProtocolMd))

}
