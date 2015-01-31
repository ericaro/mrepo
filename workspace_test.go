package mrepo

import (
	"bytes"
	"fmt"
	"strings"
)

//ExampleWorkspace_ReadFrom test several version of a .sbr that all should be readable
// and generate the same "normalized" version.
func ExampleWorkspace_ReadFrom() {
	w := Workspace{wd: "wd"} // this is local repo, doesn't matter

	//this is a normalize sample of .sbr
	normalized := `"src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git"
"mdev"
"src/github.com/ericaro/mrepo_dev" "git@github.com:ericaro/mrepo.git"
"src/github.com/ericaro/mrepo_dev2" "git@github.com:ericaro/mrepo2.git"
"mprod"
"src/github.com/ericaro/mrepo_prod" "git@github.com:ericaro/mrepo.git"
"src/github.com/ericaro/mrepoa" "git@github.com:ericaro/mrepoa.git"
`
	//this is a human-edited version, that should be identical
	humanized := `
"src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git"
"toto"
"src/github.com/ericaro/mrepo_prod" "git@github.com:ericaro/mrepo.git" "mprod"
"src/github.com/ericaro/mrepo_dev" "git@github.com:ericaro/mrepo.git" "mdev"
"src/github.com/ericaro/mrepo_dev2" "git@github.com:ericaro/mrepo2.git" "mdev"
"mprod"
"src/github.com/ericaro/mrepoa" "git@github.com:ericaro/mrepoa.git"
`
	//and this is a human edited legacy version (order is messy)
	legacy := `
"src/github.com/ericaro/mrepo" "git@github.com:ericaro/mrepo.git" "master"
"src/github.com/ericaro/mrepo_prod" "git@github.com:ericaro/mrepo.git" "mprod"
"src/github.com/ericaro/mrepo_dev" "git@github.com:ericaro/mrepo.git" "mdev"
"src/github.com/ericaro/mrepo_dev2" "git@github.com:ericaro/mrepo2.git" "mdev"
"src/github.com/ericaro/mrepoa" "git@github.com:ericaro/mrepoa.git" "mprod"
`
	// all shall give the same normalize form
	normalizer := func(sbr string) string {
		sbrs, _ := w.ReadFrom(strings.NewReader(sbr))
		buf := new(bytes.Buffer)
		WriteSubrepositoryTo(buf, sbrs)
		return buf.String()
	}

	if n := normalizer(normalized); n != normalized {
		fmt.Printf("normalization should be invariant:\n%q\n%q\n", n, normalized)
	}
	if n := normalizer(humanized); n != normalized {
		fmt.Printf("humanized should lead to same result:\n%q\n%q\n", n, normalized)
	}
	if n := normalizer(legacy); n != normalized {
		fmt.Printf("humanized should lead to same result:\n%q\n%q\n", n, normalized)
	}
	fmt.Printf("Ok\n")

	//Output: Ok
}
