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
	// ci daemon to start a daemon (righ now right here)
	// ci dashboard (to start a dashboard righ now right here)
	// ci add (uses git config ingo for addr, name, and local git remote for remote)
	// ci rem ( uses local git config ro remove jobname)
	return fs
}
