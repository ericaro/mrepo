package cmd

import (
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

type StatusCmd struct{}

func (c *StatusCmd) Run(args []string) {

	//get the revision to compare to (defaulted to origin/master)
	branch := "origin/master"
	if len(args) == 1 {
		branch = args[0]
	}

	//creates a workspace to be able to read from/to sets
	workspace, err := sbr.FindWorkspace(os.Getwd())
	if err != nil {
		exit(-1, "%v", err)
	}

	all := make([]string, 0, 100)
	//get all path, and sort them in alpha order
	all = workspace.ScanRel()

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
		left, right, giterr := git.RevListCountHead(x, branch)
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

		//and print
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", l, r, rel, errmess)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%v\t%v\t%s\t \n", tLeft, tRight, "Total")
	w.Flush()
	fmt.Println()
}
