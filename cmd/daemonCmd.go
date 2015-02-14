package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ericaro/mrepo/ci"
)

type DaemonCmd struct {
	dbfile   *string
	port     *int
	hookport *int
	hook     *bool
}

func (c *DaemonCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.dbfile = fs.String("o", "ci.db", "override the default local file name")
	c.port = fs.Int("p", 2020, "override the default local port")
	c.hookport = fs.Int("hp", 2121, "override the default hook port ")
	c.hook = fs.Bool("hook", false, "also start an http Hook server (Get returns a status, Post fire a build)")

	return fs
}

func (c *DaemonCmd) Run(args []string) {

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	daemon, err := ci.NewDaemon(wd, *c.dbfile)
	if err != nil {
		fmt.Printf("Cannot create Daemon %v", err)
		os.Exit(-1)
	}
	if *c.hook {
		go func() {
			hook := ci.NewHookServer(daemon)
			log.Printf("Starting Hook Server:%v", *c.hookport)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *c.hookport), hook))
		}()
	}

	//create a protobuf server for this daemon
	pbs := ci.NewProtobufServer(daemon)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *c.port), pbs))

}
