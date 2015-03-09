package sbr

import "fmt"

type Delta struct {
	Old Sub
	New Sub
}

//Empty return true if both old and new are equals
func (d *Delta) Empty() bool {
	return d.Old == d.New
}
func (d Delta) Rel() string {
	return d.Old.Rel()
}
func (s Delta) String() string {
	return fmt.Sprintf("%v\t%v\t%v", s.diff(s.Old.Rel(), s.New.Rel()), s.diff(s.Old.Remote(), s.New.Remote()), s.diff(s.Old.Branch(), s.New.Branch()))
}

func (d *Delta) diff(src string, target string) (res string) {
	if target == src {
		return src
	}
	return fmt.Sprintf("%s→%s", src, target)
}

//Apply changes described in 'x' to d
func Patch(d *Sub, delta Delta) (changed bool, err error) {
	// changed and err are passed on to the field patcher func
	fpatcher(&d.rel, delta.Old.rel, delta.New.rel, &changed, &err)
	fpatcher(&d.remote, delta.Old.remote, delta.New.remote, &changed, &err)
	fpatcher(&d.branch, delta.Old.branch, delta.New.branch, &changed, &err)
	return

}

//fpatcher change the actual string, if from -> to is different
// checks optimistic lock (from must be equal to actual)
// update changed, and err pointer accordingly
func fpatcher(actual *string, from, to string, changed *bool, err *error) {
	if err != nil || from == to || *err != nil {
		return
	}
	if actual == nil {
		*err = fmt.Errorf("cannot patch a nil")
		return
	}
	// now actual is not nil
	if *actual != from {
		*err = fmt.Errorf("patch optimistic lock actual <> from")
		return
	}

	*changed, *actual = true, to //apply it

}
