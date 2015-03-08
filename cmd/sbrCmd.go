package cmd

import (
	"github.com/ericaro/mrepo/format"

	"github.com/ericaro/command"
	"github.com/ericaro/help"
)

type SbrCmd struct {
	command.Commander
}

func NewSbrCmd() SbrCmd {

	c := SbrCmd{command.New()}
	c.On("clone", "<remote> [path]", "clone a remote repository, and then checkout it's .sbr", &CloneCmd{})
	c.On("checkout", "", "pull top; clone new dependencies; pull all other dependencies (deprecated dependencies can be pruned using -f option)", &CheckoutCmd{})
	c.On("version", "", "compute the sha1 of all dependencies' sha1", &VersionCmd{})
	//these are edits
	c.On("diff", "", "list subrepositories to be added to or removed from '.sbr'", &DiffCmd{})

	// utils
	c.On("x", "<command> <args>", "exec arbitrary command on each subrepository", &ExecCmd{})
	c.On("status", "[revision]", "count commits between HEAD and 'revision'", &StatusCmd{})
	c.On("format", " ", "rewrite current '.sbr' into a cannonical format", &FormatCmd{})

	// CI subcommands
	ci := command.New()
	c.On("ci", "<command> <args>", "remote ci commander. Type 'sbr ci' for help", ci)
	ci.On("log", "", "print remote log", &CilogCmd{})
	ci.On("subscribe", "", "subscribe this repository into the remote CI", &SubscribeCmd{})
	ci.On("serve", "", "start a remote CI server", &DaemonCmd{})
	ci.On("dashboard", "", "start a Dashboard web app, to display the ci server.", &DashboardCmd{})

	//also declare docs
	c.On("help", "[sections...]", "display sections summary, or section details", help.Command)
	help.Section("format", "sbr format description", SbrFormatMd)
	help.Section("ci", "CI server manual", CIServerMd)
	help.Section("protocol", "CI server protocol", string(format.CIProtocolMd))

	return c

}
