package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type WriteCmd struct {
}

func (c *WriteCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *WriteCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	del, ins, upd := workspace.WorkingDirPatches()
	//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
	// so we permute the assignmeent
	// therefore del are subrepo to be deleted from disk
	// the output will be fully tabbed

	//read ".sbr" content
	current := workspace.FileSubrepositories()

	current.RemoveAll(del)
	current.AddAll(ins)
	current.UpdateAll(upd)

	w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
	fmt.Fprintf(w, ".sbr\tpath\tremote\tbranch\n")
	for _, s := range del {
		fmt.Fprintf(w, "\033[00;32mDEL\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
	}
	for _, s := range ins {
		fmt.Fprintf(w, "\033[00;31mINS\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
	}
	for _, s := range upd {
		fmt.Fprintf(w, "\033[00;34mUPD\033[00m\t%s\t%s\t%s\t\n", diff(s.Rel(), s.XRel()), diff(s.Remote(), s.XRemote()), diff(s.Branch(), s.XBranch()))
	}
	w.Flush()
	//always rewrite the file
	WriteSbr(workspace, current)
	fmt.Printf("Done (\033[00;32m%v\033[00m INS) (\033[00;32m%v\033[00m DEL)\n", len(ins), len(del))
}
