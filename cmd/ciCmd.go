package cmd

import (
	"github.com/ericaro/command"

	"flag"
)

type CICmd struct {
	command.Commander
}

func (c *CICmd) Flags(fs *flag.FlagSet) {
	c.On("log", "", "print remote log", &CilogCmd{})
	c.On("subscribe", "", "subscribe this repository into the remote CI", &SubscribeCmd{})
	c.On("serve", "", "start a remote CI server", &DaemonCmd{})
	c.On("dashboard", "", "start a Dashboard web app, to display the ci server.", &DashboardCmd{})
}
