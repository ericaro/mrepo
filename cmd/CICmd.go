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
	//TODO(EA): add smart commands like:
	// ci daemon: start a ci daemon  ( no dependencies on .sbr)
	// ci dashboard: start a ci dashboard server (needs a remote addr)
	return fs
}
