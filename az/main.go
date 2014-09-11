package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/mrepo"
	"os"
)

// the main that run a command on all sub commands

var asynch = flag.Bool("a", false, "Controls the execution mode.\n           '-a' or '-a=true' run commands asynchronously.\n           '-a=false' of by default run commands sequentially.")
var list = flag.Bool("l", false, "Dry mode just list the repositories.")

var vcs_git = flag.Bool("git", false, "Run <command> only on git repository.")
var vcs_hg = flag.Bool("hg", false, "Run <command> only on hg  repository.")
var vcs_bzr = flag.Bool("bzr", false, "Run <command> only on bzr repository.")
var vcs_svn = flag.Bool("svn", false, "Run <command> only on svn repository.")
var vcs_cvs = flag.Bool("cvs", false, "Run <command> only on cvs repository.")

var help = flag.Bool("h", false, "Print this help.")

func main() {
	flag.Parse()
	if flag.NArg() == 0 || *help {
		fmt.Printf(`USAGE %s [-options] <command> <args...>
			
DESCRIPTION:

  Run '<command> <args...>' is every repository found in the current directory hierarchy.

OPTIONS:
	
`, os.Args[0])
		flag.PrintDefaults()

		fmt.Println("\nEXAMPLE:\n")

		fmt.Printf("'%s git status -s'\n", os.Args[0])
		os.Exit(-1)
	}

	vcs := parseVcs()
	if vcs == 0 {
		vcs = mrepo.All
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
	}

	args := make([]string, 0)
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}
	name := flag.Arg(0)

	scanner := mrepo.NewScan(wd, vcs)
	go func() {

		err = scanner.Find()
		if err != nil {
			fmt.Printf("Error scanning current directory (%s). %s", wd, err.Error())
		}
	}()

	//get the repository chan generated by Find call.
	repositories := scanner.Repositories()

	if *list {
		mrepo.List(repositories)
	} else {

		// based on the async option, exec asynchronously or sequentially.
		// we cannot just make "seq" a special case of concurrent, since when running sequentially we provide
		// direct access to the std streams. commands can use stdin, and use term escape codes.
		// When in async mode, we just can't do that.
		if *asynch {
			mrepo.Concurrent(repositories, name, args...)
		} else {
			mrepo.Seq(repositories, name, args...)
		}
	}

}

//parseVcs convert the various options into a VCS int
func parseVcs() (vcs mrepo.VCS) {

	if *vcs_git {
		vcs |= mrepo.Git
	}
	if *vcs_hg {
		vcs |= mrepo.Mercurial
	}
	if *vcs_bzr {
		vcs |= mrepo.Bazaar
	}
	if *vcs_svn {
		vcs |= mrepo.Subversion
	}
	if *vcs_cvs {
		vcs |= mrepo.Cvs
	}
	return
}
