package git

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

/* from man git

   Interrogation commands
       git-cat-file(1)
           Provide content or type and size information for repository objects.

       git-diff-files(1)
           Compares files in the working tree and the index.

       git-diff-index(1)
           Compare a tree to the working tree or index.

       git-diff-tree(1)
           Compares the content and mode of blobs found via two tree objects.

       git-for-each-ref(1)
           Output information on each ref.

       git-ls-files(1)
           Show information about files in the index and the working tree.

       git-ls-remote(1)
           List references in a remote repository.

       git-ls-tree(1)
           List the contents of a tree object.

       git-merge-base(1)
           Find as good common ancestors as possible for a merge.

       git-name-rev(1)
           Find symbolic names for given revs.

       git-pack-redundant(1)
           Find redundant pack files.

       git-rev-list(1)
           Lists commit objects in reverse chronological order.

       git-show-index(1)
           Show packed archive index.

       git-show-ref(1)
           List references in a local repository.

       git-unpack-file(1)
           Creates a temporary file with a blob’s contents.

       git-var(1)
           Show a Git logical variable.

       git-verify-pack(1)
           Validate packed Git archive files.

       In general, the interrogate commands do not touch the files in the working tree.

*/

/*
all command we are interested in (short term)
git-for-each-ref
git-ls-remote

// maybe in the future
git-cat-file(1)
git-diff-files(1)
git-diff-index(1)
git-diff-tree(1)
git-ls-files(1)
git-ls-tree(1)
git-merge-base(1)
git-name-rev(1)
git-pack-redundant(1)
git-rev-list(1)
git-show-index(1)
git-show-ref(1)
git-unpack-file(1)
git-var(1)
git-verify-pack(1)

*/

const (
	BlobType = ObjectType(iota)
	TreeType
	CommitType
	TagType
)

var (
	ObjectTypeByName = map[string]ObjectType{
		"blob":   BlobType,
		"tree":   TreeType,
		"commit": CommitType,
		"tag":    TagType,
	}
)

type ObjectType int

//ForEachRef reads all refs from local git.
func ForEachRef(prj string) (refs []Ref, err error) {
	fmt := `"%(refname)","%(objecttype)"`
	cmd := exec.Command("git", "for-each-ref", "--format", fmt)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	r := csv.NewReader(bytes.NewReader(out))
	r.FieldsPerRecord = 2
	alls, err := r.ReadAll()
	if err != nil {
		return
	}
	result := make([]Ref, len(alls))
	for i, rec := range alls {
		result[i] = Ref{
			Name: rec[0],
			Type: ObjectTypeByName[rec[1]],
		}
	}

	return result, nil

}

type Ref struct {
	Type ObjectType
	Name string
}

func ConfigGet(prj, key string) (string, error) {

	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	result := strings.Trim(string(out), DefaultTrimCut)
	return result, nil

}
func ConfigAdd(prj, key, val string) error {

	cmd := exec.Command("git", "config", "--add", key, val)
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v : %s", err, string(out))
	}
	result := strings.Trim(string(out), DefaultTrimCut)

	if result != "" {
		return errors.New(result)
	}
	return nil

}
