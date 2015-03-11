package cmd

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ericaro/sbr/sbr"
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

	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(CodeNoWorkingDir, "%v", err)
	}

	if *c.dry {

		// compute patches
		dirs, err := workspace.Scan()
		if err != nil {
			exit(CodeNoWorkingDir, "%v", err)
		}
		sbrs, err := workspace.Read()
		if err != nil {
			exit(CodeNoWorkingDir, "%v", err)
		}
		ins, del, upd := sbr.Diff(dirs, sbrs)

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
				fmt.Fprintf(w, "\033[00;34mCHANGED\033[00m\t%s\t%s\t%s\t\n", s.String())
			}
			w.Flush()
		}
		return
	}

	if *c.prune {
		fmt.Printf("PRUNE mode\n")
	}
	ch := sbr.NewCheckouter(workspace, os.Stdout)
	ch.SetPrune(*c.prune)
	ch.SetFastForwardOnly(*c.ffonly)
	ch.SetRebase(*c.rebase)

	_, err = ch.Checkout()
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
