package main

import (
	"os"

	"github.com/ericaro/command"
	"github.com/ericaro/sbr/cmd"
)

func main() {
	command.Launch(cmd.NewSbrCmd(), os.Args[0], os.Args)
}
