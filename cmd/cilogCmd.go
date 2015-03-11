package cmd

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ericaro/sbr/format"
)

type CilogCmd struct {
	tail *bool
	// TODO add server and jobname optional
}

func (c *CilogCmd) Flags(fs *flag.FlagSet) {
	c.tail = fs.Bool("tail", false, "print the current job, and then poll for updates.")
}
func (c *CilogCmd) Run(args []string) {
	wd := FindRootCmd()

	server, jobname := GetCIConf(wd)
	log.Printf("sending log request %s %s", server, jobname)

	req := &format.Request{
		Log: &format.LogRequest{
			Jobname: &jobname,
		},
	}
	b, r := GetRemoteExecution(server, req)

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

			newb, newr := GetRemoteExecution(server, req)

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
