package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type CompareCmd struct {
}

func (c *CompareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *CompareCmd) Run(args []string) {
	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	del, ins, upd := workspace.WorkingDirPatches()
	//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
	// so we permute the assignmeent
	// therefore del are subrepo to be deleted from disk
	// the output will be fully tabbed

	if len(del)+len(ins)+len(upd) > 0 {

		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', tabwriter.AlignRight)
		fmt.Fprintf(w, "\033[00;31mOPS\033[00m\tpath\tremote\tbranch\t\n")
		for _, s := range del {
			fmt.Fprintf(w, "\033[00;32mDEL\033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
		}
		for _, s := range ins {
			fmt.Fprintf(w, "\033[00;31mINS\033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
		}
		for _, s := range upd {
			fmt.Fprintf(w, "\033[00;34mUPD\033[00m\t%s\t%s\t%s\t\n", diff(s.Rel(), s.XRel()), diff(s.Remote(), s.XRemote()), diff(s.Branch(), s.XBranch()))
		}
		w.Flush()
	}
}

func diff(src string, target *string) (res string) {
	if target == nil {
		return src
	}
	// f := ansifmt.Format{}
	// f.SetStrike(true)
	// old := f.Coder()

	return fmt.Sprintf("%s→%s", (src), *target)

}
