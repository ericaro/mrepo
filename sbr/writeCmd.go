package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type writeCmd struct {
}

func (c *writeCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *writeCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	del, ins := workspace.WorkingDirPatches()
	//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
	// so we permute the assignmeent
	// therefore del are subrepo to be deleted from disk
	// the output will be fully tabbed

	//read ".sbr" content
	current := workspace.FileSubrepositories()

	current.RemoveAll(del)
	current.AddAll(ins)

	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
	fmt.Fprintf(w, ".sbr\tpath\tremote\tbranch\n")
	for _, s := range del {
		fmt.Fprintf(w, "\033[00;32mDEL\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
	}
	for _, s := range ins {
		fmt.Fprintf(w, "\033[00;31mINS\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
	}
	w.Flush()
	//always rewrite the file
	workspace.WriteSubrepositoryFile(current)
	fmt.Printf("Done (\033[00;32m%v\033[00m INS) (\033[00;32m%v\033[00m DEL)\n", len(ins), len(del))
}
