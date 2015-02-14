package cmd

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"sort"

	"github.com/ericaro/mrepo"
	"github.com/ericaro/mrepo/git"
)

type VersionCmd struct{}

func (c *VersionCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }

func (c *VersionCmd) Run(args []string) {
	// use wd by default
	wd := FindRootCmd()
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)

	all := make([]string, 0, 100)
	//get all path, and sort them in alpha order
	for _, x := range workspace.WorkingDirSubpath() {
		all = append(all, x)
	}

	sort.Sort(byName(all))

	// now compute the sha1
	h := sha1.New()
	for _, x := range all {
		// compute the sha1 for x
		version, err := git.RevParseHead(x)
		if err != nil {
			fmt.Printf("invalid subrepository, cannot compute current sha1: %s", err.Error())
		} else {
			fmt.Fprint(h, version)
		}
	}

	v := h.Sum(nil)
	fmt.Printf("%x\n", v)
}

//byName to sort any slice of Execution by their Name !
type byName []string

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i] < a[j] }
