package sbr

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {

	s1 := New("1", "r1", "b1")
	s2 := New("2", "r2", "b2")
	s3 := New("3", "r3", "b3")
	s4 := New("4", "r4", "b4")
	s5 := New("5", "r5", "b5")
	s1p := New("1", "r1", "b1p")

	src := []Sub{s1, s2, s3, s4}
	dest := []Sub{s1p, s2, s4, s5}
	ins, del, upd := Diff(src, dest)

	xins, xdel, xupd := []Sub{s5}, []Sub{s3}, []Delta{Delta{s1, s1p}}
	if !Equals(ins, xins) {
		t.Errorf("ins should be equals: %v vs %v", ins, xins)
	}
	if !Equals(del, xdel) {
		t.Errorf("del should be equals: %v vs %v", del, xdel)
	}

	if len(upd) != 1 || xupd[0] != upd[0] {
		t.Errorf("upd should be equals: %v vs %v", upd, xupd)
	}

}

func TestEquals(t *testing.T) {

	s1 := New("1", "r1", "b1")
	s2 := New("2", "r2", "b2")
	s3 := New("3", "r3", "b3")
	s4 := New("4", "r4", "b4")
	s5 := New("5", "r5", "b5")
	var res, x []Sub

	res = []Sub{s1, s2, s3, s4, s5}
	x = []Sub{s1, s2, s3, s4, s5}
	if !Equals(x, res) {
		t.Errorf("should be equals: %v vs %v", res, x)
	}

	res = []Sub{}
	x = []Sub{}
	if !Equals(x, res) {
		t.Errorf("should be equals: %v vs %v", res, x)
	}

	res = []Sub{s1, s2, s3, s4}
	x = []Sub{s1, s2, s3, s4, s5}
	if Equals(x, res) {
		t.Errorf("should be different: %v vs %v", res, x)
	}

	res = []Sub{s1, s2, s3, s4, s5}
	x = []Sub{s1, s2, s4, s3, s5}
	if Equals(x, res) {
		t.Errorf("should be different: %v vs %v", res, x)
	}

}

func TestRemoveAll(t *testing.T) {

	s1 := New("1", "r1", "b1")
	s2 := New("2", "r2", "b2")
	s3 := New("3", "r3", "b3")
	s4 := New("4", "r4", "b4")
	s5 := New("5", "r5", "b5")
	sources := []Sub{s1, s2, s3, s4, s5}

	var res, x []Sub

	x = []Sub{s1, s2, s3, s4}
	res, _ = RemoveAll(sources, s5)
	if !Equals(res, x) {
		t.Errorf("remove failed last: %v vs %v", res, x)
	}
	x = []Sub{s2, s3, s4, s5}
	res, _ = RemoveAll(sources, s1)
	if !Equals(res, x) {
		t.Errorf("remove failed first: %v vs %v", res, x)
	}

	x = []Sub{}
	res, _ = RemoveAll(sources, s1, s2, s3, s4, s5)
	if !Equals(res, x) {
		t.Errorf("remove failed All: %v vs %v", res, x)
	}
}

//ExampleWorkspace_ReadFrom test several version of a .sbr that all should be readable
// and generate the same "normalized" version.
func ExampleReadFrom() {
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
		sbrs, _ := ReadFrom(strings.NewReader(sbr))
		buf := new(bytes.Buffer)
		WriteTo(buf, sbrs)
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
