package sbr

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ericaro/sbr/git"
)

//Checkouter holds all methods to change the workspace content
type Checkouter struct {
	wk                    *Workspace
	w                     io.Writer
	prune, ffonly, rebase bool

	cloned map[string]bool // map of cloned path (to avoid pull them again)
}

//NewCheckouter creates a checkouter.
// logs are reported into w
func NewCheckouter(workspace *Workspace, w io.Writer) *Checkouter {
	return &Checkouter{
		wk: workspace,
		w:  w,
	}
}

func (c *Checkouter) SetPrune(prune bool)         { c.prune = prune }
func (c *Checkouter) SetFastForwardOnly(ffo bool) { c.ffonly = ffo }
func (c *Checkouter) SetRebase(rebase bool)       { c.rebase = rebase }

//Checkout the current workspace
//
// it's a Pull Top
// and clone/prune/pull each subrepositories
//
//
func (ch *Checkouter) Checkout() (digest []byte, err error) {

	var refresherrors []error // we keep track of all errors, but we still go on.

	err = ch.PullTop()
	if err != nil {
		return nil, err
	}

	ch.cloned, err = ch.patchDisk()
	if err != nil {
		return nil, err
	}

	// struct is ok ! update all
	err = ch.PullAll()
	if err != nil {
		refresherrors = append(refresherrors, err)
	}

	fmt.Fprintf(ch.w, "\n")

	// now compute the sha1 of all sha1
	//
	v, err := ch.wk.Version()
	if err != nil {
		fmt.Fprintf(ch.w, "ERR  Getting Version %q\n", err.Error())
		refresherrors = append(refresherrors, err)
	}
	fmt.Fprintf(ch.w, "Workspace Version %x\n", v)
	if len(refresherrors) > 0 {
		//TODO(EA) if len(errors) not too big print them out too
		return v, fmt.Errorf("Errors occured (%v) during operations", len(refresherrors))
	}
	return v, nil

}

//PullTop launches a git pull --ff-only on the Wd top git
func (ch *Checkouter) PullTop() (err error) {
	result, err := git.Pull(ch.wk.Wd(), ch.ffonly, ch.rebase)
	if err != nil {
		fmt.Fprintf(ch.w, "ERR  Pulling '/'   : %q\n%s\n", err.Error(), result)
		return
	}
	fmt.Fprintf(ch.w, "     Pulling '/'...\n")
	return
}

func (ch *Checkouter) PullAll() (err error) {
	var waiter sync.WaitGroup

	for _, prj := range ch.wk.ScanRel() {
		if !ch.cloned[prj] { //

			waiter.Add(1)
			go func(prj string) {
				defer waiter.Done()
				res, e := git.Pull(prj, ch.ffonly, ch.rebase)
				if e != nil {
					fmt.Fprintf(ch.w, "ERR  Pulling '%s'   : %q\n%s\n", prj, e.Error(), res)
					if err == nil {
						err = e
					}

				} else {
					fmt.Fprintf(ch.w, "     Pulling '%s'...\n", prj)
				}
			}(prj)
		}
	}
	waiter.Wait() // wait all Pulls

	return nil
}

//applychanges computes ins, del, upd , and try to apply them on the workspace
func (ch *Checkouter) patchDisk() (cloned map[string]bool, err error) {
	wds, err := ch.wk.Scan()
	if err != nil {
		return
	}
	sbrs, err := ch.wk.Read()
	if err != nil {
		return
	}

	ins, del, upd := Diff(wds, sbrs)

	// map to keep track of cloned repo (that don't need refresh)
	cloned = make(map[string]bool)

	var refresherrors []error // we keep track of all errors, but we still go on.

	var waiter sync.WaitGroup // to wait for all commands to return
	var delCount, cloneCount, changeCount int

	if len(ins) > 0 || len(del) > 0 || len(upd) > 0 {

		for _, sbr := range ins {
			waiter.Add(1)
			go func(d Sub) {
				defer waiter.Done()
				res, err := git.Clone(ch.wk.Wd(), d.Rel(), d.Remote(), d.Branch())
				if err != nil {
					fmt.Fprintf(ch.w, "ERR  Cloning into '%s'   : %q\n%s\n", d.Rel(), err.Error(), res)
					refresherrors = append(refresherrors, err)
				} else {
					cloneCount++
					fmt.Fprintf(ch.w, "     Cloning into '%s'...\n", d.Rel())
				}
			}(sbr)
		}

		for _, delta := range upd {
			u, err := ch.UpdateRepository(delta)
			if err != nil {
				fmt.Fprintf(ch.w, "ERR  Changing '%s'   : %s\n%s\n", delta.Rel(), err.Error(), delta.String())
				refresherrors = append(refresherrors, err)
			} else {
				if u {
					fmt.Fprintf(ch.w, "     Changing %s\n", delta.String())
					changeCount++
				}
			}

		}

		if ch.prune {

			for _, sbr := range del {
				waiter.Add(1)
				go func(d Sub) {
					defer waiter.Done()
					err = ch.Prune(sbr)
					if err != nil {
						fmt.Fprintf(ch.w, "ERR  Pruning '%s'   : %q\n", d.Rel(), err.Error())
						refresherrors = append(refresherrors, err)
					} else {
						delCount++
						fmt.Fprintf(ch.w, "     Pruning '%s'...\n", d.Rel())
					}
				}(sbr)
			}
		} // no prune at all

		waiter.Wait()
		//print out some report
		if ch.prune {
			fmt.Fprintf(ch.w, "%v CLONE, %v PRUNE %v CHANGED\n\n", cloneCount, delCount, changeCount)
		} else {
			//fake prune with a specific message
			for _, sbr := range del {
				fmt.Fprintf(ch.w, "     Would Prune %s %s %s\n", sbr.Rel(), sbr.Remote(), sbr.Branch())
				delCount++
			}
			fmt.Fprintf(ch.w, "%v CLONE, %v REQUIRED PRUNE %v CHANGED\n\n", cloneCount, delCount, changeCount)
		}
	}
	if len(refresherrors) > 0 {
		// todo print those errors if there are not too many
		err = fmt.Errorf("Errors occured (%v) during operations", len(refresherrors))
	}
	return
}

//locate return the absolute path of a rel path
func (ch *Checkouter) locate(rel string) string {
	return filepath.Join(ch.wk.Wd(), rel)
}

//Prune a Sub
func (ch *Checkouter) Prune(d Sub) (err error) {
	path := ch.locate(d.rel)
	_, err = os.Stat(path)
	if os.IsNotExist(err) { // it does not exists
		return nil
	}
	return os.RemoveAll(filepath.Join(ch.wk.Wd(), d.rel))
}

//Update a repository according to changes described in delta
func (ch *Checkouter) UpdateRepository(delta Delta) (updated bool, err error) {

	u, err := ch.UpdateBranch(delta)
	if err != nil {
		return
	}
	updated = updated || u

	u, err = ch.UpdateRemote(delta)
	if err != nil {
		return
	}
	updated = updated || u
	return
}

//Update branch on actual git repo if needed
func (ch *Checkouter) UpdateBranch(delta Delta) (updated bool, err error) {

	path := ch.locate(delta.Old.rel)

	oldbranch, err := git.Branch(path)
	if err != nil {
		return
	}
	branch := delta.New.Branch()

	if branch == oldbranch {
		return false, nil // nothing to do
	}

	// we need to update
	exists, err := git.BranchExists(path, branch)
	if err != nil {
		return false, err
	}
	err = git.Checkout(path, branch, !exists)
	if err != nil {
		return false, err
	}
	return true, err
}

//Update remote updates the remote origin
func (ch *Checkouter) UpdateRemote(delta Delta) (updated bool, err error) {

	path := ch.locate(delta.Old.rel)

	oldremote, err := git.RemoteOrigin(path)
	if err != nil {
		return
	}
	remote := delta.New.Remote()

	if remote == oldremote {
		return false, nil // nothing to do
	}

	err = git.RemoteSetOrigin(path, remote)
	if err != nil {
		return
	}
	return true, nil

}
