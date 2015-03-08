package mrepo

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/ericaro/mrepo/git"
)

var (
	//LegacyFmt set a global legacy format writing
	LegacyFmt = false
)

//Workspace represent the current workspace.
//
// It mainly deal with:
//
// - reading working dir for subrepositories
//
// - reading/writing .sbr file for subrepositories
//
//
type Workspace struct {
	wd          string          //current working dir
	sbrfilename string          // the .sbr filename (by default .sbr)
	fileSbr     Subrepositories // subrepositories as declared in the .sbr
	wdSbr       Subrepositories // subrepositories as found in the workspace.
}

//NewWorkspace creates a new Workspace for a working dir.
func NewWorkspace(wd string) *Workspace {
	return &Workspace{
		wd:          wd,
		sbrfilename: ".sbr",
	}
}

//Wd return the current working directory for this workspace.
func (x *Workspace) Wd() string {
	return x.wd
}

//Sbrfile return the workspace sbr file name.
func (x *Workspace) Sbrfile() string {
	return filepath.Join(x.wd, x.sbrfilename)
}

//WorkingDirDependencies scan once the working dir for subrepositories
func (x *Workspace) WorkingDirSubrepositories() Subrepositories {

	if x.wdSbr == nil { // lazy part

		wdSbr := make(Subrepositories, 0, 100)

		for _, prj := range x.WorkingDirSubpath() {
			branch, err := git.Branch(prj)
			if err != nil {
				log.Fatalf("%s doesn't seem to have branches: %s", prj, err.Error())
			}
			origin, err := git.RemoteOrigin(prj)
			if err != nil {
				log.Fatalf("%s doesn't declare a remote 'origin': %s", prj, err.Error())
			}
			rel := x.Relativize(prj)
			if rel != "." {
				wdSbr = append(wdSbr, Subrepository{
					wd:     x.wd,
					rel:    rel,
					remote: origin,
					branch: branch,
				})
			}
		}
		sort.Sort(wdSbr)
		x.wdSbr = wdSbr
	}
	return x.wdSbr
}

//FileSubrepositories returns a set of Subrepositories, as declared in the .sbr file
func (x *Workspace) FileSubrepositories() (wdSbr Subrepositories) {
	if x.wdSbr == nil {
		file, err := os.Open(filepath.Join(x.wd, x.sbrfilename))
		if err == nil {
			defer file.Close()
			x.fileSbr, err = x.ReadFrom(file) // for now, just parse
			if err != nil {

				fmt.Printf("invalid dependency file %q format. Skipping\n", x.sbrfilename)
			}

		} else {
			if os.IsNotExist(err) {
				fmt.Printf("dependency file %q does not exists. Skipping\n", x.sbrfilename)
			} else {
				fmt.Printf("Error reading dependency file %q: %s", x.sbrfilename, err.Error())
			}
		}
	}
	return x.fileSbr
}

func WriteSbr(file io.Writer, sbr Subrepositories) {

	sort.Sort(sbr)
	pbranch := "master" // the previous branch : init to default

	for _, d := range sbr {
		if LegacyFmt {
			fmt.Fprintf(file, "git %q %q %q\n", d.rel, d.remote, d.branch)
		} else {

			if d.branch != pbranch {
				//declare new branch section
				fmt.Fprintf(file, "%q\n", d.branch)
			}

			fmt.Fprintf(file, "%q %q\n", d.rel, d.remote)
			pbranch = d.branch
		}
	}
}

// //WorkingDirPatches computes changes to be applied to the working dir
// func (w *Workspace) WorkingDirPatches() (ins, del Subrepositories, upd []XSubrepository) {
// 	target := w.FileSubrepositories()
// 	src := w.WorkingDirSubrepositories()
// 	ins, del, upd = src.Diff(target)
// 	return
// }

// //WorkingDirPatches computes changes to be applied to the sbrfile
// func (w *Workspace) SbrfilePatches() (ins, del Subrepositories, upd []XSubrepository) {
// 	target := w.WorkingDirSubrepositories()
// 	src := w.FileSubrepositories()
// 	ins, del, upd = src.Diff(target)
// 	return
// }

//WorkingDirSubpath extract only the path of the subrepositories (faster than the whole dependency)
func (x *Workspace) WorkingDirSubpath() []string {
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

//relpath computes the relative path of a subrepository
func (x *Workspace) Relativize(subrepository string) string {
	if filepath.IsAbs(subrepository) {
		rel, err := filepath.Rel(x.wd, subrepository)
		if err != nil {
			fmt.Printf("prj does not appear to be in the current directory %s\n", err.Error())
		}
		return rel
	}
	return subrepository
}

//ReadFrom read subrepository definitions fom reader
func (p *Workspace) ReadFrom(r io.Reader) (sbr Subrepositories, err error) {

	w := csv.NewReader(r)
	w.Comma = ' '
	w.FieldsPerRecord = -1 // allow variable fields
	w.Comment = '#'

	records, err := w.ReadAll()
	if err != nil {
		return
	}
	sbr = make([]Subrepository, 0, len(records)) // not the real size but a good approx of the "size"

	// closure to make "newSubrepository" as easy as it looks
	newSubrepository := func(path, remote, branch string) {
		r := Subrepository{
			wd:     p.wd,
			rel:    path,
			remote: remote,
			branch: branch,
		}
		sbr = append(sbr, r)

	}

	currentBranch := "master"
	for i, record := range records {
		switch len(record) {
		case 1:
			currentBranch = record[0]
		case 2:
			newSubrepository(record[0], record[1], currentBranch)
		case 3:
			log.Printf("Warning: Subrepository %q format is not normalized. use 'sbr format' to fix it.", record[0])
			newSubrepository(record[0], record[1], record[2])
		case 4: //legacy
			log.Printf("Warning: Subrepository %q uses legacy format. use 'sbr format' to fix it.", record[1])
			newSubrepository(record[1], record[2], record[3])
		default:
			err = fmt.Errorf("invalid %vth record #fields must be 1,2,3, or 4 not %v", i, len(record))
			return
		}
	}
	return
}
