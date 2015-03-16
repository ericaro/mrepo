package cmd

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ericaro/sbr/dashboard"
	"github.com/ericaro/sbr/git"
)

type DashboardCmd struct {
	server *string
	title  *string
	port   *int
	prop   *float64
}

func (c *DashboardCmd) Flags(fs *flag.FlagSet) {
	c.server = fs.String("s", "", "remote server address")
	c.title = fs.String("t", "CI Dashboard", "CI title")
	c.port = fs.Int("p", 8080, "http port to listen to")
	c.prop = fs.Float64("prop", 4, "cell width ~= prop*cell height")
}

func (c *DashboardCmd) Run(args []string) {

	//use the server from git config if none were specified
	if *c.server == "" { // no one specified use the one in config get
		var err error
		wd := FindRootCmd()
		*c.server, err = git.ConfigGet(wd, "ci.server")
		if err != nil {
			fmt.Printf("Error, cannot read ci remote address in git config. %s\n Use \ngit config --add ci.server <ci address>\n", err.Error())
			os.Exit(CodeMissingServerConfig)
		}
	}

	log.Printf("dashboard available at http://locahost:%v displaying %s \n", *c.port, *c.server)
	d := new(dashboard.Dashboard)
	d.Title = *c.title
	d.Prop = *c.prop
	d.Server = *c.server
	http.ListenAndServe(fmt.Sprintf(":%v", *c.port), d)
}
