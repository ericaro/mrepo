package mrepo

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ericaro/mrepo/git"
)

var (
	ErrNotYetSupported = errors.New("Not yet Supported")
)

//XSubrepository contains information to "update" a subrepository (kind of diff)
//
// it aggregates a Subrepository (the origin) and the differences.
//
// Differences are just "pointer" to the original value (here *string)
//
// nil means no differences
// a value is the new value.
type XSubrepository struct {
	Subrepository         //
	wd            *string // absolute path for the working dir
	rel           *string //relative path for the project
	remote        *string
	branch        *string
}

func NewXSubrepository(src, target Subrepository) XSubrepository {
	return XSubrepository{
		Subrepository: src,
		wd:            diffString(src.wd, target.wd),
		rel:           diffString(src.rel, target.rel),
		remote:        diffString(src.remote, target.remote),
		branch:        diffString(src.branch, target.branch),
	}

}

//Empty return true when there is no difference between src and target
func (x *XSubrepository) Empty() bool {
	return x.wd == nil && x.rel == nil && x.remote == nil && x.branch == nil
}

//String  returns a diff representation
func (x *XSubrepository) String() string {
	return fmt.Sprintf("%s %s %s",
		stringdiff(x.Rel(), x.rel),
		stringdiff(x.Remote(), x.remote),
		stringdiff(x.Branch(), x.branch),
	)
}
func (x *XSubrepository) XRel() *string    { return x.rel }
func (x *XSubrepository) XRemote() *string { return x.remote }
func (x *XSubrepository) XBranch() *string { return x.branch }

// the idea is to provide "apply" methods that can apply  on the disk, changes described by this xsbr.
// all changes cannot be applied: wd, cannot, rel, hardly (not sure I would want that anyway, because rel is used to identify stuff)
// on the other hand, remote, and branch should be easy to.

// func (x *XSubrepository) UpdateWorkingDir() (updated bool, err error) {
// 	return false, ErrNotYetSupported
// }
// func (x *XSubrepository) UpdatePath() (updated bool, err error) {
// 	return false, ErrNotYetSupported
// }

func (x *XSubrepository) Update() (updated bool, err error) {
	u, err := x.UpdateBranch()
	if err != nil {
		return
	}
	updated = updated || u

	u, err = x.Updateremote()
	if err != nil {
		return
	}
	updated = updated || u
	return
}

//Update branch on actual git repo if needed
func (x *XSubrepository) UpdateBranch() (updated bool, err error) {
	if x.branch == nil {
		return false, nil // nothing to do
	}
	// we need to update
	prj := filepath.Join(x.Subrepository.wd, x.Rel())
	branch := *x.branch

	exists, err := git.BranchExists(prj, branch)
	// log.Printf("updating branch %s (exists? %v)", x.branch, exists)
	if err != nil {
		return false, err
	}
	err = git.Checkout(prj, branch, !exists)
	if err != nil {
		return false, err
	}
	return true, err
}

func (x *XSubrepository) Updateremote() (updated bool, err error) {
	if x.remote == nil {
		return false, nil
	}
	return false, ErrNotYetSupported
}

func diffString(src, target string) *string {
	if src == target {
		return nil
	}
	return &target
}

func stringdiff(src string, target *string) (res string) {
	if target == nil {
		return src
	}

	return fmt.Sprintf("%s→%s", src, *target)

}
