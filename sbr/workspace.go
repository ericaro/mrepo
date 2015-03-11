package sbr

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ericaro/sbr/git"
)

const (
	SbrFile = ".sbr"
)

var (
	ErrNoSbrfile = errors.New("Not in an 'sbr' workspace")
)

type Workspace struct {
	wd       string //current working dir
	filename string // the .sbr filename (by default .sbr)
}

//NewWorkspace creates a new Workspace for a working dir.
func NewWorkspace(wd string) *Workspace {
	return &Workspace{
		wd:       wd,
		filename: SbrFile,
	}
}

//FindWorkspace get the current working dir and search for a .sbr file upwards
//
// FindWorkspace(os.Getwd() )
func FindWorkspace(root string, oserr error) (w *Workspace, err error) {
	if oserr != nil {
		return nil, oserr
	}
	path := root
	//loop until I've reached the root, or found the .sbr
	for ; !fileExists(filepath.Join(path, SbrFile)) && path != "/"; path = filepath.Dir(path) {
	}

	if path != "/" {
		return NewWorkspace(path), nil
	} else {
		return NewWorkspace(root), ErrNoSbrfile
	}
}

//Sbrfile return the workspace sbr file name.
func (x *Workspace) Sbrfile() string { return filepath.Join(x.wd, x.filename) }

//Wd return the current working directory for this workspace.
func (x *Workspace) Wd() string { return x.wd }

//Read returns the []Sub, as declared in the .sbr file
func (x *Workspace) Read() (sbrs []Sub, err error) {

	file, err := os.Open(x.Sbrfile())
	if err != nil {
		return
	}
	defer file.Close()
	return ReadFrom(file) // for now, just parse
}

//Scan the working dir and return subrepositories found
func (x *Workspace) Scan() (sbrs []Sub, err error) {

	sbrs = make([]Sub, 0, 100)

	for _, prj := range x.ScanRel() {
		// this is a git repo, read all three fields
		branch, err := git.Branch(prj)
		if err != nil {
			return sbrs, fmt.Errorf("%s doesn't have branches: %s", prj, err.Error())
		}
		origin, err := git.RemoteOrigin(prj)
		if err != nil {
			return sbrs, fmt.Errorf("%s doesn't declare a remote 'origin': %s", prj, err.Error())
		}
		rel, err := filepath.Rel(x.wd, prj)
		if err != nil {
			return sbrs, fmt.Errorf("%s not in the current directory %s\n", err.Error())
		}
		if rel != "." {
			sbrs = append(sbrs, New(rel, origin, branch))
		}
	}
	Sort(sbrs)
	return
}

//ScanRel extract only the path of the subrepositories (faster than the whole dependency)
func (x *Workspace) ScanRel() []string {
	prjc := make([]string, 0, 100)
	//the subrepo scanner function
	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if f.Name() == ".git" {
				// it's a repository file
				prjc = append(prjc, filepath.Dir(path))
				//always skip the repository file
				return filepath.SkipDir
			}
		}
		return nil
	}
	filepath.Walk(x.wd, walker)
	return prjc
}

//Version compute the workspace version (the sha1 of all sha1)
func (wk *Workspace) Version() (version []byte, err error) {
	//get all path, and sort them in alpha order
	subs := wk.ScanRel()
	all := make([]string, 0, len(subs))
	errs := make([]string, 0, len(subs))
	for _, x := range subs {
		all = append(all, x)
	}

	sort.Strings(all)

	// now compute the sha1
	h := sha1.New()
	for _, x := range all {
		// compute the sha1 for x
		version, err := git.RevParseHead(x)
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprint(h, version)
		}
	}
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "\n"))
		return
	}

	v := h.Sum(nil)
	return v, nil
}

//fileExists check if a path exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
