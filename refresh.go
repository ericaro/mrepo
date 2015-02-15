package mrepo

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/ericaro/mrepo/git"
)

var (
	ErrRefreshAll = errors.New("error refreshing subrepositories")
)

//PullTop launches a git pull --ff-only on the Wd top git
func (wk *Workspace) PullTop(w io.Writer, ffonly, rebase bool) (err error) {
	result, err := git.Pull(wk.Wd(), ffonly, rebase)
	if err != nil {
		fmt.Fprintf(w, "ERR  Pulling '/'   : %q\n%s\n", err.Error(), result)
		return
	}
	fmt.Fprintf(w, "     Pulling '/'...\n")
	return
}
func (wk *Workspace) PullAll(w io.Writer, skip map[string]bool, ffonly, rebase bool) (err error) {
	var waiter sync.WaitGroup

	for _, prj := range wk.WorkingDirSubpath() {
		// it would be nice to git pull in async mode, really.
		if !skip[prj] { //

			waiter.Add(1)
			go func(prj string) {
				defer waiter.Done()
				res, e := git.Pull(prj, ffonly, rebase)
				if e != nil {
					fmt.Fprintf(w, "ERR  Pulling '%s'   : %q\n%s\n", prj, e.Error(), res)
					if err == nil {
						err = e
					}

				} else {
					fmt.Fprintf(w, "     Pulling '%s'...\n", prj)
				}
			}(prj)
		}
	}
	waiter.Wait() // wait all Pulls

	return nil
}

func (wk *Workspace) Checkout(w io.Writer, prune, ffonly, rebase bool) (digest []byte, err error) {

	var refresherrors []error // we keep track of all errors, but we still go on.

	err = wk.PullTop(w, ffonly, rebase)
	if err != nil {
		return nil, err
	}

	cloned, err := wk.ApplyChanges(w, prune)
	if err != nil {
		return nil, err
	}

	// struct is ok ! update all
	err = wk.PullAll(w, cloned, ffonly, rebase)
	if err != nil {
		refresherrors = append(refresherrors, err)
	}

	fmt.Fprintf(w, "\n")

	// now compute the sha1 of all sha1
	//
	v, err := wk.Version()
	if err != nil {
		fmt.Fprintf(w, "ERR  Getting Version %q\n", err.Error())
		refresherrors = append(refresherrors, err)
	}
	fmt.Fprintf(w, "Workspace Version %x\n", v)
	if len(refresherrors) > 0 {
		//TODO(EA) if len(errors) not too big print them out too
		return v, fmt.Errorf("Errors occured (%v) during operations", len(refresherrors))
	}
	return v, nil

}
func (wk *Workspace) ApplyChanges(w io.Writer, prune bool) (cloned map[string]bool, err error) {

	ins, del, upd := wk.WorkingDirPatches()
	if err != nil {
		return
	}
	// map to keep track of cloned repo (that don't need refresh)
	cloned = make(map[string]bool)

	var refresherrors []error // we keep track of all errors, but we still go on.

	var waiter sync.WaitGroup // to wait for all commands to return
	var delCount, cloneCount, changeCount int

	if len(ins) > 0 || len(del) > 0 || len(upd) > 0 {

		for _, sbr := range ins {
			waiter.Add(1)
			go func(d Subrepository) {
				defer waiter.Done()
				res, err := d.Clone()
				if err != nil {
					fmt.Fprintf(w, "ERR  Cloning into '%s'   : %q\n%s\n", d.Rel(), err.Error(), res)
					refresherrors = append(refresherrors, err)
				} else {
					cloneCount++
					fmt.Fprintf(w, "     Cloning into '%s'...\n", d.Rel())
				}
			}(sbr)
		}
		for _, xsbr := range upd {
			u, err := xsbr.Update()
			if err != nil {
				fmt.Fprintf(w, "ERR  Changing '%s'   : %s\n%s\n", xsbr.Rel(), err.Error(), xsbr.String())
				refresherrors = append(refresherrors, err)
			} else {
				if u {
					fmt.Fprintf(w, "     Changing %s\n", xsbr.String())
					changeCount++
				}
			}

		}

		if prune {

			for _, sbr := range del {
				waiter.Add(1)
				go func(d Subrepository) {
					defer waiter.Done()
					err = sbr.Prune()
					if err != nil {
						fmt.Fprintf(w, "ERR  Pruning '%s'   : %q\n", d.Rel(), err.Error())
						refresherrors = append(refresherrors, err)
					} else {
						delCount++
						fmt.Fprintf(w, "     Pruning '%s'...\n", d.Rel())
					}
				}(sbr)
			}
		} // no prune at all

		waiter.Wait()

		// after all, if prune was false just print out the prune
		if prune {
			fmt.Fprintf(w, "%v CLONE, %v PRUNE %v CHANGED\n\n", cloneCount, delCount, changeCount)
		} else {
			for _, sbr := range del {
				fmt.Fprintf(w, "     Would Prune %s %s %s\n", sbr.Rel(), sbr.Remote(), sbr.Branch())
				delCount++
			}
			fmt.Fprintf(w, "%v CLONE, %v REQUIRED PRUNE %v CHANGED\n\n", cloneCount, delCount, changeCount)
		}
	}
	if len(refresherrors) > 0 {
		err = fmt.Errorf("Errors occured (%v) during operations", len(refresherrors))
	}
	return
}

//Version compute the workspace version (the sha1 of all sha1)
func (wk *Workspace) Version() (version []byte, err error) {
	//get all path, and sort them in alpha order
	subs := wk.WorkingDirSubpath()
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
