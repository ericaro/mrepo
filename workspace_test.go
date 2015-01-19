package mrepo

import (
	"bytes"
	"fmt"
	"strings"
)

func ExampleWorkspace_parseDependencies() {
	w := Workspace{wd: "wd"}
	r := strings.NewReader(`git "rel" "remote" "dev"
git "rel2" "remote2"
git "rel3" "remote3" "prod"
`)
	deps := w.parseDependencies(r)
	buf := new(bytes.Buffer)
	WriteSubrepositoryTo(buf, deps)
	fmt.Println(buf.String())

	//Output:
	// git "rel" "remote" "dev"
	// git "rel2" "remote2"
	// git "rel3" "remote3" "prod"
}
func ExampleWorkspace_parseDependencies_default() {
	w := Workspace{wd: "wd"}
	r := strings.NewReader(`git "path" "remote"
git "path2" "remote3"
git "path3" "remote3"
`)
	deps := w.parseDependencies(r)
	buf := new(bytes.Buffer)
	fmt.Printf("branch %s\n", deps[0].Branch())
	WriteSubrepositoryTo(buf, deps)

	//Output:
	// branch master
}
