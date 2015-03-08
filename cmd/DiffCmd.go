package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type DiffCmd struct {
	apply, meld *bool
}

func (d *DiffCmd) Flags(fs *flag.FlagSet) {
	d.apply = fs.Bool("apply", false, "if true update the '.sbr' with the changes")
	d.meld = fs.Bool("meld", false, "Use Meld to display differences (implies -apply==false)")
}

func (d *DiffCmd) Run(args []string) {

	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	if *d.meld {
		//generate a temp file
		current := workspace.WorkingDirSubrepositories()
		f, err := ioutil.TempFile("", "sbr")
		if err != nil {
			fmt.Printf("Cannot generate temp file: %s", err.Error())
			os.Exit(-1)

		}
		mrepo.WriteSbr(f, current)
		f.Close() //no defer to open it up just after.
		err = mrepo.Meld(workspace.Wd(), ".sbr set  |  disk set", workspace.Sbrfile(), f.Name())
		if err != nil {
			fmt.Printf("Meld returned with error: %s", err.Error())
			os.Exit(-1)
		}

		return
	}

	//compute patches (to be made to the working dir)
	current := workspace.FileSubrepositories()
	dest := workspace.WorkingDirSubrepositories()
	ins, del, upd := mrepo.Diff(current, dest)
	//print them
	if len(del)+len(ins)+len(upd) > 0 {

		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', tabwriter.AlignRight)
		fmt.Fprintf(w, "\033[00;31mOPS \033[00m\tpath\tremote\tbranch\t\n")
		for _, s := range upd {
			fmt.Fprintf(w, "\033[00;34mEDIT\033[00m\t%s\t%s\t%s\t\n", d.diffR(s.Rel(), s.XRel()), d.diffR(s.Remote(), s.XRemote()), d.diffR(s.Branch(), s.XBranch()))
		}
		for _, s := range ins {
			fmt.Fprintf(w, "\033[00;31mADD \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
		}
		for _, s := range del {
			fmt.Fprintf(w, "\033[00;32mDEL \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
		}
		w.Flush()
	}

	if !*d.apply { // end of the road, below we actually apply changes to .sbr
		return
	}

	//del, ins, upd := workspace.WorkingDirPatches()
	//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
	// so we permute the assignmeent
	// therefore del are subrepo to be deleted from disk
	// the output will be fully tabbed

	//read ".sbr" content
	//current := workspace.FileSubrepositories()

	current.RemoveAll(del)
	current.AddAll(ins)
	current.UpdateAll(upd)
	//always rewrite the file
	WriteSbr(workspace, current)
	fmt.Printf("Done (\033[00;32m%v\033[00m INS) (\033[00;32m%v\033[00m DEL) (\033[00;32m%v\033[00m UPD)\n", len(ins), len(del), len(upd))
}

//present reverse changes
func (c *DiffCmd) diffR(src string, target *string) (res string) {
	if target == nil {
		return src
	}
	return fmt.Sprintf("%s→%s", *target, src)
}
