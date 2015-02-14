package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type CompareCmd struct {
	reverse *bool
}

func (c *CompareCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.reverse = fs.Bool("r", false, "with present changes as .sbr file edition.")
	return fs
}

func (c *CompareCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	ins, del, upd := workspace.WorkingDirPatches()
	//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
	// so we permute the assignmeent
	// therefore del are subrepo to be deleted from disk
	// the output will be fully tabbed
	if *c.reverse {

		if len(del)+len(ins)+len(upd) > 0 {

			w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', tabwriter.AlignRight)
			fmt.Fprintf(w, "\033[00;31mOPS \033[00m\tpath\tremote\tbranch\t\n")
			for _, s := range ins {
				fmt.Fprintf(w, "\033[00;32mDEL \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range del {
				fmt.Fprintf(w, "\033[00;31mADD \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range upd {
				fmt.Fprintf(w, "\033[00;34mEDIT\033[00m\t%s\t%s\t%s\t\n", diffR(s.Rel(), s.XRel()), diffR(s.Remote(), s.XRemote()), diffR(s.Branch(), s.XBranch()))
			}
			w.Flush()
		}

	} else {

		if len(del)+len(ins)+len(upd) > 0 {

			w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', tabwriter.AlignRight)
			fmt.Fprintf(w, "\033[00;31mOPS   \033[00m\tpath\tremote\tbranch\t\n")
			for _, s := range del {
				fmt.Fprintf(w, "\033[00;32mPRUNE \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range ins {
				fmt.Fprintf(w, "\033[00;31mCLONE \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range upd {
				fmt.Fprintf(w, "\033[00;34mUPDATE\033[00m\t%s\t%s\t%s\t\n", diff(s.Rel(), s.XRel()), diff(s.Remote(), s.XRemote()), diff(s.Branch(), s.XBranch()))
			}
			w.Flush()
		}
	}
}

//present changes to be made to the right
func diff(src string, target *string) (res string) {
	if target == nil {
		return src
	}
	return fmt.Sprintf("%s→%s", (src), *target)
}

//present reverse changes
func diffR(src string, target *string) (res string) {
	if target == nil {
		return src
	}
	return fmt.Sprintf("%s→%s", *target, src)
}
