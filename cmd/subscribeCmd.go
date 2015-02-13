package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/ericaro/ci"

	"github.com/ericaro/mrepo/git"
)

const (
	CodeNoWorkingDir        = -2
	CodeMissingServerConfig = -3
	CodeMissingJobConfig    = -4
	CodeCannotDelete        = -5
	CodeMissingBranch       = -6
	CodeMissingRemoteOrigin = -7
	CodeCannotAddJob        = -8
)

type SubscribeCmd struct {
	force  *bool
	remove *bool
	server *string
	// TODO add server and jobname optional
}

func (c *SubscribeCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.force = fs.Bool("force", false, "force remote job creation even if it exists")
	c.remove = fs.Bool("remove", false, "only remove remote job.")
	c.server = fs.String("server", "", "the remote address ci adress")
	return fs
}
func (c *SubscribeCmd) Run(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error, cannot determine the current directory. %s\n", err.Error())
		os.Exit(CodeNoWorkingDir)
	}

	if *c.server == "" { // no one specified use the one in config get

		*c.server, err = git.ConfigGet(wd, "ci.server")
		if err != nil {
			fmt.Printf("Error, cannot read ci remote address in git config. %s\n Use \ngit config --add ci.server <ci address>\n", err.Error())
			os.Exit(CodeMissingServerConfig)
		}
	} else { // there was one, store it
		err := git.ConfigAdd(wd, "ci.server", *c.server)
		if err != nil {
			fmt.Printf("Warning, cannot add ci remote address in git config. %v\n", err)
		}
		fmt.Printf("git config' key \"ci.server\" = %s\n", *c.server)
	}
	cl := ci.NewClient(*c.server)

	var jobname string
	if len(args) == 0 { // no job name specified, use the one in config

		jobname, err = git.ConfigGet(wd, "ci.job.name")
		if err != nil {
			fmt.Printf("Error, cannot read ci remote job name in git config. %s\n Use \ngit config --add ci.job.name <jobname>\n", err.Error())
			os.Exit(CodeMissingJobConfig)
		}

	} else { // there was one job name store it
		jobname = args[0]
		git.ConfigAdd(wd, "ci.job.name", jobname)
		if err != nil {
			fmt.Printf("Warning, cannot add ci remote job name in git config. %v\n", err)
		}
		fmt.Printf("git config' key \"ci.job.name\" = %s\n", jobname)
	}

	if *c.remove { // with only remove I don't need to go further

		cl.RemoveJob(jobname)
		if err != nil {
			fmt.Printf("Error, cannot delete remote job %s. %v\n", jobname, err)
			os.Exit(CodeCannotDelete)
		}
	}

	branch, err := git.Branch(wd)
	if err != nil {
		fmt.Printf("Error, cannot get local branch. %v\n", err)
		os.Exit(CodeMissingBranch)
	}

	remote, err := git.RemoteOrigin(wd)
	if err != nil {
		fmt.Printf("Error, cannot get remote/origin url. %v\n", err)
		os.Exit(CodeMissingRemoteOrigin)
	}
	// I now have server job name, and branch, they have been stored for later use

	if *c.force { // needs a delete first
		err := cl.RemoveJob(jobname) // don't care about the error
		if err != nil {
			fmt.Printf("Warning: deletion error: %v", err)
		}
	}
	err = cl.AddJob(jobname, remote, branch)
	if err != nil {
		fmt.Println(err)
		os.Exit(CodeCannotAddJob)
	}
	// notify success
	fmt.Printf("added %s %s %s\n", jobname, remote, branch)
}
