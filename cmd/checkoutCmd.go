package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
)

type CheckoutCmd struct {
	prune  *bool
	ffonly *bool
	rebase *bool
	dry    *bool
}

func (c *CheckoutCmd) Flags(fs *flag.FlagSet) {
	c.prune = fs.Bool("prune", false, "prune sub repositories that are not in the .sbr file")
	c.ffonly = fs.Bool("ff-only", false, "Refuse to merge and exit with a non-zero status unless the current HEAD is already up-to-date or the merge can be resolved as a fast-forward.")
	c.rebase = fs.Bool("rebase", false, "rebase instead of merge")
	c.dry = fs.Bool("d", false, "dry run. Only print out what would be applied")
}

func (c *CheckoutCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	if *c.dry {

		// compute patches
		ins, del, upd := mrepo.Diff(workspace.WorkingDirSubrepositories(), workspace.FileSubrepositories())

		if len(del)+len(ins)+len(upd) > 0 {

			w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', tabwriter.AlignRight)
			fmt.Fprintf(w, "\033[00;31mOPS    \033[00m\tpath\tremote\tbranch\t\n")
			for _, s := range del {
				fmt.Fprintf(w, "\033[00;32mPRUNE  \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range ins {
				fmt.Fprintf(w, "\033[00;31mCLONE  \033[00m\t%s\t%s\t%s\t\n", s.Rel(), s.Remote(), s.Branch())
			}
			for _, s := range upd {
				fmt.Fprintf(w, "\033[00;34mCHANGED\033[00m\t%s\t%s\t%s\t\n", c.diff(s.Rel(), s.XRel()), c.diff(s.Remote(), s.XRemote()), c.diff(s.Branch(), s.XBranch()))
			}
			w.Flush()
		}
		return
	}

	if *c.prune {
		fmt.Printf("PRUNE mode\n")
	}
	_, err := workspace.Checkout(os.Stdout, *c.prune, *c.ffonly, *c.rebase)
	if err != nil {
		fmt.Printf("checkout error: %s", err.Error())
		os.Exit(-1)
	}

}

//present changes to be made to the right
func (c *CheckoutCmd) diff(src string, target *string) (res string) {
	if target == nil {
		return src
	}
	return fmt.Sprintf("%s→%s", (src), *target)
}
