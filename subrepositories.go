package mrepo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ericaro/mrepo/git"
)

//this files contains functions that deals with subrepositories

//Subrepository type contains all the information about a subrepository.
type Subrepository struct {
	wd     string // absolute path for the working dir
	rel    string //relative path for the project
	remote string
	branch string
}

//copy returns a copy of this subrepository
func (d Subrepository) copy() Subrepository {
	return Subrepository{
		wd:     d.wd,
		rel:    d.rel,
		remote: d.remote,
		branch: d.branch,
	}
}

//Rel returns this project's relative path.
func (d *Subrepository) Rel() string {
	return d.rel
}

//Remote returns this project's remote.
func (d *Subrepository) Remote() string {
	return d.remote
}

//Branch returns this project's branch.
func (d *Subrepository) Branch() string {
	return d.branch
}
func (d *Subrepository) String() string {
	return fmt.Sprintf("%s %s %s", d.rel, d.remote, d.branch)
}

//Apply changes described in 'x' to d
func (d *Subrepository) Apply(x XSubrepository) (changed bool, err error) {
	// always the same pattern if x.ss != nil mean there is a patch to do
	// if d.ss != x.Subrepository.ss means optimistic lock
	if x.wd != nil {
		if d.wd != x.Subrepository.wd {
			err = errors.New("Conflict wd")
		}
		d.wd = *x.wd
		changed = true
	}
	if x.rel != nil {
		if d.rel != x.Subrepository.rel {
			err = errors.New("Conflict rel")
		}
		d.rel = *x.rel
		changed = true
	}
	if x.remote != nil {
		if d.remote != x.Subrepository.remote {
			err = errors.New("Conflict remote")
		}
		d.remote = *x.remote
		changed = true
	}
	if x.branch != nil {
		if d.branch != x.Subrepository.branch {
			err = fmt.Errorf("Conflict branch was %v!=%v → %v", d.branch, x.Subrepository.branch, *x.branch)
		}
		d.branch = *x.branch
		changed = true
	}
	return

}

func (d *Subrepository) Exists() (exists bool, err error) {
	_, err = os.Stat(filepath.Join(d.wd, d.rel))
	if os.IsNotExist(err) { // I need to create one
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *Subrepository) Clone() (result string, err error) {
	exists, err := d.Exists()
	if err != nil {
		return "", fmt.Errorf("cannot test %s : %s", filepath.Join(d.wd, d.rel), err.Error())
	}
	if !exists {
		return git.Clone(d.wd, d.rel, d.remote, d.branch)
	}
	return "Ok", nil
}

func (d *Subrepository) Prune() (err error) {
	path := filepath.Join(d.wd, d.rel)
	_, err = os.Stat(path)
	if os.IsNotExist(err) { // it does not exists
		return nil
	}
	return os.RemoveAll(filepath.Join(d.wd, d.rel))
}

//Subrepositories represent a set of subrepositories.
// Subrepositories are always stored sorted by "rel"
type Subrepositories []Subrepository

func (a Subrepositories) Len() int      { return len(a) }
func (a Subrepositories) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Subrepositories) Less(i, j int) bool {
	//sortint subrepository, first key is branch, second is rel
	if a[i].branch == a[j].branch {
		return a[i].rel < a[j].rel
	} else {
		return a[i].branch < a[j].branch
	}
}

//AddAll append a bunch of subrepositories to 'd'
func (d *Subrepositories) AddAll(ins Subrepositories) (changed bool) {
	sources := *d
	for _, d := range ins {
		sources = append(sources, d)
		changed = true
	}
	*d = sources
	return
}
func (d *Subrepositories) UpdateAll(upd []XSubrepository) (changed bool) {
	sources := *d
	ix := make(map[string]XSubrepository) // make it a map for fast (and cleaner) queries
	for _, upd := range upd {
		ix[upd.Subrepository.Rel()] = upd
	}

	for i, sbr := range sources {
		x, exists := ix[sbr.Rel()]
		if exists {
			u, err := sources[i].Apply(x)
			if err != nil {
				log.Printf("Conflict During patch: %s", err)
			}
			changed = changed || u
		}
	}
	return
}

//RemoveAll subrepositories from 'd'
func (d *Subrepositories) RemoveAll(del Subrepositories) (changed bool) {
	sources := *d
	deleted := indexSbr(del)
	j := 0
	for i, d := range sources {
		if _, del := deleted[d.rel]; !del { // we simply copy the values, deletion is just an offset in fact
			if i != j { // if apply is false, then sources will never be changed
				sources[j] = sources[i]
				changed = true
			}
			j++
		}
	}
	*d = sources[0:j]
	return
}

//Diff compute the changes to be applied to 'src', in order to became dest.
// updates are not handled, just insertion, and deletion.
//later, maybe we'll add update for branches
func Diff(src, dest Subrepositories) (insertion, deletion []Subrepository, update []XSubrepository) {
	//TODO  add a map[string]struct{Old,New} to see updates
	ins, del, upd := make([]Subrepository, 0, 100), make([]Subrepository, 0, 100), make([]XSubrepository, 0, 100)
	//give a identifying string  for each sbr, then, I will only have to met the differences.
	targets := indexSbr(dest)
	currents := indexSbr(src)

	//then compute the diffs
	for id, t := range targets { // for each dest
		_, exists := currents[id]
		if !exists { // if missing , create an insert
			ins = append(ins, t)
		}
	}
	for id, c := range currents { // for each src
		_, exists := targets[id]
		if !exists { // locally exists, but not in dest, it's a deletion
			del = append(del, c)
		}
	}
	//compute the upd
	for id, src := range currents { // for each src
		dest, exists := targets[id]
		if exists {
			x := NewXSubrepository(src, dest)
			if !x.Empty() {
				upd = append(upd, x)
			}
		}

	}

	return ins, del, upd
}

//indexSbr build up a small index of Subrepository based on their .rel attribute.
func indexSbr(deps []Subrepository) map[string]Subrepository {
	i := make(map[string]Subrepository, 100)
	for _, x := range deps {
		i[x.rel] = x
	}
	return i
}
