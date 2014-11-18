package mrepo

import (
	"fmt"
	"os"
	"path/filepath"
)

//this files contains functions that deals with subrepositories

//Subrepository type contains all the information about a subrepository.
type Subrepository struct {
	wd     string // absolute path for the working dir
	rel    string //relative path for the project
	remote string
	branch string
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
	return fmt.Sprintf("git %q %q %q", d.rel, d.remote, d.branch)
}

func (d *Subrepository) Clone() (result string, err error) {
	//trry to stat the directory
	_, err = os.Stat(filepath.Join(d.wd, d.rel))
	if os.IsNotExist(err) { // I need to create one
		return GitClone(d.wd, d.rel, d.remote, d.branch)
	} else {
		return "", fmt.Errorf("cannot clone into %s, destination path already exists.")
	}
}

func (d *Subrepository) Prune() (err error) {
	path := filepath.Join(d.wd, d.rel)
	_, err = os.Stat(path)
	if os.IsNotExist(err) { // it does not exists
		return nil
	} else {
		return os.RemoveAll(filepath.Join(d.wd, d.rel))
	}
}

//Subrepositories represent a set of subrepositories.
// Subrepositories are always stored sorted by "rel"
type Subrepositories []Subrepository

func (a Subrepositories) Len() int           { return len(a) }
func (a Subrepositories) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Subrepositories) Less(i, j int) bool { return a[i].rel < a[j].rel }

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

//Remove subrepositories from 'd'
// apply make this method act like a dry run
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

//Diff compute the changes to be applied to 'current', in order to became target.
// updates are not handled, just insertion, and deletion.
//later, maybe we'll add update for branches
func (current Subrepositories) Diff(target Subrepositories) (insertion, deletion Subrepositories) {
	ins, del := make([]Subrepository, 0, 100), make([]Subrepository, 0, 100)
	targets := indexSbr(target)
	currents := indexSbr(current)

	//then compute the diffs
	for id, t := range targets { // for each target
		_, exists := currents[id]
		if !exists { // if missing , create an insert
			ins = append(ins, t)
		}
	}
	for id, c := range currents { // for each current
		_, exists := targets[id]
		if !exists { // locally exists, but not in target, it's a deletion
			del = append(del, c)
		}
	}
	return ins, del
}

//indexSbr build up a small index of Subrepository based on their .rel attribute.
func indexSbr(deps []Subrepository) map[string]Subrepository {
	i := make(map[string]Subrepository, 100)
	for _, x := range deps {
		i[x.rel] = x
	}
	return i
}
