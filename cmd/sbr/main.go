package main

import (
	"os"

	"github.com/ericaro/command"
	"github.com/ericaro/mrepo/cmd"
)

func main() {
	command.Launch(cmd.NewSbrCmd(), os.Args[0], os.Args)
}
