package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ericaro/mrepo/git"

	"github.com/ericaro/ci/cmd"
	"github.com/ericaro/ci/format"
)

type CilogCmd struct {
	tail *bool
	// TODO add server and jobname optional
}

func (c *CilogCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.tail = fs.Bool("tail", false, "print the current job, and then poll for updates.")
	return fs
}
func (c *CilogCmd) Run(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(-1)
	}
	//creates a workspace to be able to read from/to sets
	//workspace := mrepo.NewWorkspace(wd)

	server, jobname := GetCIConf(wd)

	req := &format.Request{
		Log: &format.LogRequest{
			Jobname: &jobname,
		},
	}
	b, r := cmd.GetRemoteExecution(server, req)

	fmt.Println(r.Print(), "\n")

	if r.Done() { // if refresh has finished, print the build
		fmt.Println(r.Summary())

		fmt.Println(b.Print(), "\n")
		if b.Done() { // if build has finished, print a summary
			fmt.Println(b.Summary())
		}
	}

	if *c.tail {
		for _ = range time.Tick(2 * time.Second) {

			newb, newr := cmd.GetRemoteExecution(server, req)

			fmt.Print(r.Tail(newr))
			fmt.Print(b.Tail(newb))
			b, r = newb, newr
		}
	} else { // when not in tail mode, always print out the summary for both refresh, and build
		if b.StartAfter(r) {
			fmt.Println("\n\n", r.Summary(), "\n", b.Summary())
		} else {
			fmt.Println("\n\n", b.Summary(), "\n", r.Summary())
		}
	}
}

func GetCIConf(prj string) (server, jobname string) {
	server, err := git.ConfigGet(prj, "ci.server")
	if err != nil {
		fmt.Printf("Error, cannot read ci remote address in git config. %s\n Use \ngit config --add ci.server <ci address>\n", err.Error())
		os.Exit(-1)
	}
	jobname, err = git.ConfigGet(prj, "ci.job.name")
	if err != nil {
		fmt.Printf("Error, cannot read ci remote job name in git config. %s\n Use \ngit config --add ci.job.name <jobname>\n", err.Error())
		os.Exit(-1)
	}
	return server, jobname
}
