package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ericaro/sbr/git"
	"github.com/ericaro/sbr/sbr"
)

type StatusCmd struct {
	short *bool
}

func (c *StatusCmd) Flags(fs *flag.FlagSet) {
	c.short = fs.Bool("s", false, "print only repo that have differences")
}
func (c *StatusCmd) Run(args []string) {

	//get the revision to compare to (defaulted to origin/master)

	//creates a workspace to be able to read from/to sets
	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(-1, "%v", err)
	}

	//get all path, and sort them in alpha order
	all := workspace.ScanRel()

	//little trick to keep project sorted.
	// this is only possible when I execute sync commands.
	sort.Strings(all)

	//basically just running  git rev-list on each subrepo
	// left, right, err := git.RevListCountHead(x, branch)
	// and all the rest is "printing" stuff

	//pretty tab printer
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 2, ' ', 0)

	//we are going to print a gran total
	var tLeft, tRight int

	for _, x := range all {

		// the real deal

		left, right, giterr := git.RevListCountHead(x)
		tLeft += left
		tRight += right

		//computes the relative path (prettier to print)
		rel, err := filepath.Rel(workspace.Wd(), x)
		if err != nil {
			rel = x // rel is only use for presentation
		}

		//compute the left,right  string
		l := strconv.FormatInt(int64(left), 10)  //Defualt
		r := strconv.FormatInt(int64(right), 10) //default
		//pretty print 0 as -
		if left == 0 {
			l = "-"
		}
		if right == 0 {
			r = "-"
		}
		errmess := ""
		if giterr != nil { //pretty print err as ?
			l = "?"
			r = "?"
			errmess = strings.Replace(giterr.Error(), "\n", "; ", -1)
		}

		//equals if there is not changes
		equals := left == 0 && right == 0
		if !*c.short || !equals { // print only if required
			//and print
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", l, r, rel, errmess)
		}
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%v\t%v\t%s\t \n", tLeft, tRight, "Total")
	w.Flush()
	fmt.Println()
}
