package sbr

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
)

func Equals(src, dest []Sub) bool {
	if len(src) != len(dest) {
		return false
	}
	for i := range src {
		if src[i] != dest[i] {
			return false
		}
	}
	return true
}

func UpdateAll(subs []Sub, upd ...Delta) (changed bool, err error) {

	// map Delta by their old.Rel path
	index := indexDelta(upd)

	for i, sbr := range subs {
		delta, exists := index[sbr.Rel()] //find out the Delta
		if exists {
			u, e := Patch(&subs[i], *delta)
			if e != nil {
				log.Printf("Conflict During patch: %s", e)
				err = e
				return
			}
			changed = changed || u
		}
	}

	return
}

//RemoveAll subrepositories from 'd'
func RemoveAll(sources []Sub, del ...Sub) (res []Sub, changed bool) {
	deleted := indexSbr(del)

	res = make([]Sub, 0, len(sources)-len(del))

	for _, d := range sources {
		if _, del := deleted[d.rel]; !del { // we simply copy the values, deletion is just an offset in fact
			res = append(res, d)
			changed = true
		}
	}
	return
}

//Sort a slice of Sub according to the natural order.
//
func Sort(sbrs []Sub) { sort.Stable(byRelBranch(sbrs)) }

type byRelBranch []Sub

func (a byRelBranch) Len() int           { return len(a) }
func (a byRelBranch) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRelBranch) Less(i, j int) bool { return a[i].Less(a[j]) }

//Diff compute the changes to be applied to 'src', in order to became dest.
func Diff(src, dest []Sub) (insertion, deletion []Sub, update []Delta) {
	ins, del, upd := make([]Sub, 0, len(dest)), make([]Sub, 0, len(src)), make([]Delta, 0, max(len(src), len(dest)))

	//give a identifying string  for each sbr, then, I will only have to met the differences.
	idest := indexSbr(dest)
	isrc := indexSbr(src)

	//look for insertions
	for id, sbr := range idest { // for each dest
		_, exists := isrc[id]
		if !exists { // it does not exists in source
			ins = append(ins, *sbr)
		}
	}
	for id, sbr := range isrc { // for each src
		_, exists := idest[id]
		if !exists { // it does not exists in dest
			del = append(del, *sbr)
		}
	}

	//compute the upd
	for id, src := range isrc { // for each src
		dest, exists := idest[id]
		if exists { // it exists in both
			x := Delta{Old: *src, New: *dest}
			if !x.Empty() {
				upd = append(upd, x)
			}
		}
	}
	return ins, del, upd
}

//ReadFrom read subrepository definitions fom reader
//
// the initial currentBranch is 'master'
func ReadFrom(r io.Reader) (sbr []Sub, err error) { return ReadFromBranch("master", r) }

//ReadFromBranch read subrepository definitions from reader
func ReadFromBranch(currentBranch string, r io.Reader) (sbr []Sub, err error) {

	w := csv.NewReader(r)
	w.Comma = ' '
	w.FieldsPerRecord = -1 // allow variable fields
	w.Comment = '#'

	records, err := w.ReadAll()
	if err != nil {
		return
	}
	sbr = make([]Sub, 0, len(records)) // not the real size but a good approx of the "size"

	//currentBranch := "master"
	for i, record := range records {
		switch len(record) {
		case 1:
			currentBranch = record[0]
		case 2:
			sbr = append(sbr, New(record[0], record[1], currentBranch))
		case 3:
			log.Printf("Warning: Subrepository %q format is not normalized. use 'sbr format' to fix it.", record[0])
			sbr = append(sbr, New(record[0], record[1], record[2]))
		case 4: //legacy
			log.Printf("Warning: Subrepository %q uses legacy format. use 'sbr format' to fix it.", record[1])
			sbr = append(sbr, New(record[1], record[2], record[3]))
		default:
			err = fmt.Errorf("invalid %vth record #fields must be 1,2,3, or 4 not %v", i, len(record))
			return
		}
	}
	return
}

func WriteTo(w io.Writer, sbr []Sub) {
	Sort(sbr)

	pbranch := "master" // the previous branch : init to default

	for _, d := range sbr {
		if d.branch != pbranch {
			//declare new branch section
			fmt.Fprintf(w, "%q\n", d.branch)
		}

		fmt.Fprintf(w, "%q %q\n", d.rel, d.remote)
		pbranch = d.branch
	}
}

//indexSbr build up a small index of Subrepository based on their .rel attribute.
func indexSbr(deps []Sub) map[string]*Sub {
	index := make(map[string]*Sub, len(deps))
	for i := range deps {
		index[deps[i].rel] = &deps[i]
	}
	return index
}

//indexSbr build up a small index of Subrepository based on their .rel attribute.
func indexDelta(deltas []Delta) map[string]*Delta {
	index := make(map[string]*Delta, len(deltas))
	for i := range deltas {
		index[deltas[i].Old.rel] = &deltas[i]
	}
	return index
}

func max(a, b int) int {
	switch {
	case a >= b:
		return a
	default:
		return b
	}
}
