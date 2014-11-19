package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
)

const (
	Usage = `
USAGE %[1]s [-options] <command> <args...>
			
DESCRIPTION:

  %[1]s subrepositories manager for git dependencies.

  It manages two sets of subrepositories:
  
    - ".%[1]s": set made of subrepositories declarations in '.%[1]s' file
    - "disk": set made of actual subrepositories in the current directory hierarchy
  

  <command> can be:

    - describe: print the "disk" dependency set
    - compare : diff ".%[1]s" and "disk" sets. In the form of operations to apply to ".%[1]s" set.
    - reflect : replace ".sbr" set by "disk" one.
    - apply   : apply ".sbr" dependencies to the current working dir (prune and clone)
    - merge   : edit two sets in meld.
    - digest  : compute the sha1 of all the dependencies sha1.

OPTIONS:

`
	Example = `
EXAMPLES:

  - Print out all dependencies present in a workspace:
	  $ %[1]s describe

  - Init a .%[1]s to reflect the current working dir:
	  $ %[1]s reflect
    
  - Add a subrepository as usual, and update your .%[1]s:
      $ git clone git@github.com:ericaro/mrepo.git -b dev src/ericaro/mrepo
      $ %[1]s reflect

  - Apply .%[1]s changes to the working dir:
      $ %[1]s apply

`
)

var (
	dotmrepo = flag.String("s", ".sbr", "override default dependency filename")
	// workingdir = flag.String("wd", ".", "path to be used as working dir")
	help = flag.Bool("h", false, "Print this help.")
)

func usage() {
	fmt.Printf(Usage, os.Args[0])
	flag.PrintDefaults()
	fmt.Printf(Example, os.Args[0])
}

func main() {
	flag.Parse()

	if flag.NArg() <= 0 || *help {
		usage()
		return
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
	}
	//creates a workspace to be able to read from/to sets
	workspace := mrepo.NewWorkspace(wd)
	cmd := flag.Arg(0)
	switch {

	case cmd == "describe": // not diff mode, hence, plain local mode
		// execute query on each subrepo
		current := workspace.WorkingDirSubrepositories()
		// and just print it out
		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		for _, d := range current {
			fmt.Fprintf(w, "git\t%q\t%q\t%q\n", d.Rel(), d.Remote(), d.Branch())
		}
		w.Flush()
	case cmd == "merge":
		//generate a temp file
		current := workspace.WorkingDirSubrepositories()
		f, err := ioutil.TempFile("", "sbr")
		mrepo.WriteSubrepositoryTo(f, current)
		f.Close() //no defer to open it up just after.
		err = mrepo.Meld(workspace.Wd(), ".sbr set  |  disk set", workspace.Sbrfile(), f.Name())
		if err != nil {
			fmt.Printf("Meld returned with error: %s", err.Error())
			os.Exit(-1)
		}
		// shall I apply ?

	case cmd == "compare":

		del, ins := workspace.WorkingDirPatches()
		//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
		// so we permute the assignmeent
		// therefore del are subrepo to be deleted from disk
		// the output will be fully tabbed

		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		fmt.Fprintf(w, ".sbr\tpath\tremote\tbranch\n")
		for _, s := range del {
			fmt.Fprintf(w, "\033[00;32mDEL\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
		}
		for _, s := range ins {
			fmt.Fprintf(w, "\033[00;31mINS\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
		}
		w.Flush()

	case cmd == "digest":

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
			version, err := mrepo.GitRevParseHead(x)
			if err != nil {
				fmt.Printf("invalid subrepository, cannot compute current sha1: %s", err.Error())
			} else {
				fmt.Fprint(h, version)
			}
		}

		v := h.Sum(nil)
		fmt.Printf("%x\n", v)

	case cmd == "reflect":
		//compute ins and del in the .sbr file
		del, ins := workspace.WorkingDirPatches()
		//WorkingDirPatches > (ins, del) are for the wd, here we are interested in the reverse
		// so we permute the assignmeent
		// therefore del are subrepo to be deleted from disk
		// the output will be fully tabbed

		//read ".sbr" content
		current := workspace.FileSubrepositories()

		current.RemoveAll(del)
		current.AddAll(ins)

		w := tabwriter.NewWriter(os.Stdout, 3, 8, 3, ' ', 0)
		fmt.Fprintf(w, ".sbr\tpath\tremote\tbranch\n")
		for _, s := range del {
			fmt.Fprintf(w, "\033[00;32mDEL\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
		}
		for _, s := range ins {
			fmt.Fprintf(w, "\033[00;31mINS\033[00m\t%s\t%s\t%s\n", s.Rel(), s.Remote(), s.Branch())
		}
		w.Flush()
		//always rewrite the file
		workspace.WriteSubrepositoryFile(current)
		fmt.Printf("Done (\033[00;32m%v\033[00m INS) (\033[00;32m%v\033[00m DEL)\n", len(ins), len(del))

	case cmd == "apply":
		ins, del := workspace.WorkingDirPatches()
		var waiter sync.WaitGroup // to wait for all commands to return
		var delCount, cloneCount int
		for _, sbr := range ins {
			waiter.Add(1)
			go func(d mrepo.Subrepository) {
				defer waiter.Done()
				_, err := d.Clone()
				if err != nil {
					fmt.Printf("\033[00;31mERR\033[00m  git clone %s -b %s %s:\n     %s\n", d.Rel(), d.Remote(), d.Branch(), err.Error())
				} else {
					cloneCount++
					fmt.Printf("     Cloning into '%s'...\n", d.Rel())
				}
			}(sbr)
		}
		for _, sbr := range del {
			waiter.Add(1)
			go func(d mrepo.Subrepository) {
				defer waiter.Done()
				err = sbr.Prune()
				if err != nil {
					fmt.Printf("\033[00;31mERR\033[00m  rm -Rf %s :\n     %s\n", d.Rel(), err.Error())
				} else {
					delCount++
					fmt.Printf("     Pruning '%s'...\n", d.Rel())
				}
			}(sbr)
		}
		waiter.Wait()
		fmt.Printf("Done (\033[00;32m%v\033[00m CLONE) (\033[00;32m%v\033[00m PRUNE)\n", cloneCount, delCount)

	default:
		usage()
		return

	}
}

//byName to sort any slice of Execution by their Name !
type byName []string

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i] < a[j] }
