package mrepo

import (
	"os"
	"path/filepath"
)

const (
	All        VCS = 0
	Git        VCS = 1
	Mercurial  VCS = 2
	Bazaar     VCS = 3
	Subversion VCS = 4
	Cvs        VCS = 5
)

type VCS int

type Scanner interface {
	//find sub projects, and publish them into a chan.
	Find() error
	// get the chan where projects are published.
	// Once all project have been found, the chan is closed.
	Projects() <-chan string
}

//scanner object to scan for a directory looking for git projects.
type scanner struct {
	prjc     chan string
	wd       string
	dirnames map[string]bool
}

//NewScan creates a scanner
func NewScan(workingDir string, vcs VCS) Scanner {
	return &scanner{

		wd:   workingDir,
		prjc: make(chan string),
		dirnames: map[string]bool{
			".git": vcs == All || vcs == Git,
			".hg":  vcs == All || vcs == Mercurial,
			".bzr": vcs == All || vcs == Bazaar,
			".svn": vcs == All || vcs == Subversion,
			"CVS":  vcs == All || vcs == Cvs,
		},
	}

}

//Find starts the directory scanning, and publish project found.
func (s scanner) Find() (err error) {
	defer close(s.prjc)

	//err = filepath.Walk(s.wd, s.walkFn)
	// for backward compatibility (with 1.0.3) I can't call a method
	f := func(path string, f os.FileInfo, err error) error { return s.walkFn(path, f, err) }
	return filepath.Walk(s.wd, f)
	return
}

//Projects exposes the chan of project found as they are found.
// The chan is closed at the end.
func (s scanner) Projects() <-chan string {
	return s.prjc
}

//WaldirFn compatible
func (s scanner) walkFn(path string, f os.FileInfo, err error) error {
	// if this path is a "prj", add it.
	// it would be something like if it's .git => it's parent is

	if f.IsDir() {

		for dirname, add := range s.dirnames {
			if f.Name() == dirname {
				// it's a project file

				if add {
					s.prjc <- filepath.Dir(path)
				}
				//always skip the project file
				return filepath.SkipDir
			}

		}
	}
	return nil
}
