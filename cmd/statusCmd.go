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

	"github.com/ericaro/mrepo"
	"github.com/ericaro/mrepo/git"
)

type StatusCmd struct{}

func (c *StatusCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *StatusCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()

	//get the revision to compare to (defaulted to origin/master)
	branch := "origin/master"
	if len(args) == 1 {
		branch = args[0]
	}

	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	all := make([]string, 0, 100)
	//get all path, and sort them in alpha order
	for _, x := range workspace.WorkingDirSubpath() {
		all = append(all, x)
	}

	//little trick to keep project sorted.
	// this is only possible when I execute sync commands.
	sort.Sort(byName(all))

	//basically just running  git rev-list on each subrepo
	// left, right, err := git.RevListCountHead(x, branch)
	// and all the rest is "printing" stuff

	//pretty tab printer
	w := tabwriter.NewWriter(os.Stdout, 4, 8, 2, ' ', 0)

	//we are going to print a gran total
	var tLeft, tRight int

	for _, x := range all {

		// the real deal
		left, right, giterr := git.RevListCountHead(x, branch)
		tLeft += left
		tRight += right

		//computes the relative path (prettier to print)
		rel, err := filepath.Rel(wd, x)
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

		//and print
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", l, r, rel, errmess)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%v\t%v\t%s\t \n", tLeft, tRight, "Total")
	w.Flush()
	fmt.Println()
}
