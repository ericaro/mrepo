package cmd

import (
	"github.com/ericaro/command"

	"flag"
)

type CICmd struct {
	command.Commander
}

func (c *CICmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.Commander = command.NewCommander("ci", fs)
	c.On("log", "print remote log", &CilogCmd{}, nil)
	c.On("subscribe", "subscribe this repository into the remote CI", &SubscribeCmd{}, nil)
	c.On("serve", "start a remote CI server", &DaemonCmd{}, nil)
	c.On("dashboard", "start a Dashboard web app, to display the ci server.", &DashboardCmd{}, nil)
	return fs
}
