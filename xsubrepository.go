package mrepo

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

// the idea is to provide "apply" methods that can apply  on the disk, changes described by this xsbr.
// all changes cannot be applied: wd, cannot, rel, hardly (not sure I would want that anyway, because rel is used to identify stuff)
// on the other hand, remote, and branch should be easy to.

//Update branch on actual git repo
// func (x *XSubrepository) UpdateBranch() error {

// }

func diffString(src, target string) *string {
	if src == target {
		return nil
	}
	return &target
}
