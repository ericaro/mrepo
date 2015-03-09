package sbr

import "fmt"

//Sub type contains all the information about a sub.
type Sub struct {
	rel    string //relative path for the project
	remote string
	branch string
}

func New(rel, remote, branch string) Sub {
	return Sub{
		rel:    rel,
		remote: remote,
		branch: branch,
	}
}

//Rel returns this project's relative path.
func (d Sub) Rel() string { return d.rel }

//Less represent the natural order for Sub
// branch first then rel.
func (d Sub) Less(x Sub) bool { return d.branch < x.branch || (d.branch == x.branch && d.rel < x.rel) }

//Remote returns this project's remote.
func (d Sub) Remote() string { return d.remote }

//Branch returns this project's branch.
func (d Sub) Branch() string { return d.branch }
func (d Sub) String() string { return fmt.Sprintf("%s %s %s", d.rel, d.remote, d.branch) }
