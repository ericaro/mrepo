package main

import (
	"os"

	"github.com/ericaro/command"
	"github.com/ericaro/mrepo/cmd"
)

func main() {
	command.Exec(&cmd.SbrCmd{}, os.Args)
}
