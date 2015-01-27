package mrepo

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

//Workspace represent the current workspace.
//
//
// - executes a random command synchronously
//
// - executes a random command concurrently and pushes the result in a chan of 'Execution'
//
// - read local subrepositories
//
// - read local .sbr file

//
// Caveat, sync execution passes stdin/out to the subprocess that runs the command, so it can run in interactive mode,
// whereas async execution does not.
// async mode is required for statistical postprocessors.
type Workspace struct {
	wd          string          //current working dir
	sbrfilename string          // the .sbr filename (by default .sbr)
	fileSbr     Subrepositories // subrepositories as declared in the .sbr
	wdSbr       Subrepositories // subrepositories as found in the workspace.
}

//NewWorkspace creates a new Workspace for a working dir.
func NewWorkspace(wd string) *Workspace {
	return &Workspace{
		wd:          wd,
		sbrfilename: ".sbr",
	}
}

//Wd return the current working directory for this workspace.
func (x *Workspace) Wd() string {
	return x.wd
}
func (x *Workspace) Sbrfile() string {
	return filepath.Join(x.wd, x.sbrfilename)
}

//ExecSequentially, for each `subrepository` in the working dir, execute the  command `command` with arguments `args`.
// It passes the stdin, stdout, and stderr to the subprocess. and wait for the result, before moving to the next one.
func (x *Workspace) ExecSequentially(command string, args ...string) {
	var count int

	for _, sub := range x.WorkingDirSubpath() {
		count++
		rel := x.relpath(sub)
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", rel, command, strings.Join(args, " "))
		cmd := exec.Command(command, args...)
		cmd.Dir = sub
		cmd.Stderr, cmd.Stdout, cmd.Stdin = os.Stderr, os.Stdout, os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s %s':\n    %s\n", command, strings.Join(args, " "), err.Error())
		}
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//ExecConcurently, for each `subrepository` in the working dir, execute the command `command` with arguments `args`.
// Each command is executed in non interactive mode (no access to stdin/stdout)
func (x *Workspace) ExecConcurrently(command string, args ...string) <-chan Execution {
	executions := make(chan Execution)
	var waiter sync.WaitGroup // to wait for all commands to return
	for _, sub := range x.WorkingDirSubpath() {
		waiter.Add(1)

		go func(sub string) {
			defer waiter.Done()
			cmd := exec.Command(command, args...)
			cmd.Dir = sub
			out, err := cmd.CombinedOutput()
			if err != nil {
				return
			}
			rel := x.relpath(sub)
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", sub, command, strings.Join(args, " "))
			//executions <- head + string(out)
			result := string(out)
			result = strings.Trim(result, defaultTrimCut)
			executions <- Execution{Name: sub, Rel: rel, Cmd: command, Args: args, Result: result}
		}(sub)
	}

	go func() {
		waiter.Wait()
		close(executions)
	}()
	return executions
}

//WorkingDirDependencies returns, or lazily compute, the Subrepositories found in the working dir.
func (x *Workspace) WorkingDirSubrepositories() Subrepositories {

	if x.wdSbr == nil {

		wdSbr := make(Subrepositories, 0, 100)

		for _, prj := range x.WorkingDirSubpath() {
			branch, err := GitBranch(prj)
			if err != nil {
				log.Fatalf("%s doesn't seem to have branches: %s", prj, err.Error())
			}
			origin, err := GitRemoteOrigin(prj)
			if err != nil {
				log.Fatalf("%s doesn't declare a remote 'origin': %s", prj, err.Error())
			}
			rel := x.relpath(prj)
			if rel != "." {
				wdSbr = append(wdSbr, Subrepository{
					wd:     x.wd,
					rel:    rel,
					remote: origin,
					branch: branch,
				})
			}
		}
		sort.Sort(wdSbr)
		x.wdSbr = wdSbr
	}
	return x.wdSbr
}

//FileSubrepositories returns a set of Subrepositories, as declared in the .sbr file
func (x *Workspace) FileSubrepositories() (wdSbr Subrepositories) {
	if x.wdSbr == nil {
		file, err := os.Open(filepath.Join(x.wd, x.sbrfilename))
		if err == nil {
			defer file.Close()
			x.fileSbr = x.parseDependencies(file) // for now, just parse
		} else {
			if os.IsNotExist(err) {
				fmt.Printf("dependency file %q does not exists. Skipping\n", x.sbrfilename)
			} else {
				fmt.Printf("Error reading dependency file %q: %s", x.sbrfilename, err.Error())
			}
		}
	}
	return x.fileSbr
}

//WriteSubrepositoryFile write down the set of subrepositories into the default subrepositories file.
func (x *Workspace) WriteSubrepositoryFile(wdSbr Subrepositories) {
	f, err := os.Create(x.sbrfilename)
	if err != nil {
		fmt.Printf("Cannot write dependency file: %s", err.Error())
		return
	}
	defer f.Close()
	WriteSubrepositoryTo(f, wdSbr)
}
func WriteSubrepositoryTo(file io.Writer, wdSbr Subrepositories) {
	sort.Sort(wdSbr)
	pbranch := "master" // the previous branch : init to default
	for _, d := range wdSbr {
		if d.branch == pbranch {
			//short version
			fmt.Fprintf(file, "git %q %q\n", d.rel, d.remote)
		} else { //long one
			fmt.Fprintf(file, "git %q %q %q\n", d.rel, d.remote, d.branch)
		}
		pbranch = d.branch
	}
}

//WorkingDirPatches computes changes to be applied to the working dir
func (w *Workspace) WorkingDirPatches() (ins, del Subrepositories) {
	target := w.FileSubrepositories()
	current := w.WorkingDirSubrepositories()
	ins, del = current.Diff(target)
	return
}

//WorkingDirSubpath extract only the path of the subrepositories (faster than the whole dependency)
func (x *Workspace) WorkingDirSubpath() []string {
	prjc := make([]string, 0, 100)
	//the subrepo scanner function
	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if f.Name() == ".git" {
				// it's a repository file
				prjc = append(prjc, filepath.Dir(path))
				//always skip the repository file
				return filepath.SkipDir
			}
		}
		return nil
	}
	filepath.Walk(x.wd, walker)
	return prjc
}

//relpath computes the relative path of a subrepository
func (x *Workspace) relpath(subrepository string) string {
	if filepath.IsAbs(subrepository) {
		rel, err := filepath.Rel(x.wd, subrepository)
		if err != nil {
			fmt.Printf("prj does not appear to be in the current directory %s\n", err.Error())
		}
		return rel
	}
	return subrepository
}

//parseDependencies scans 'r' and fill []Subrepository
func (p *Workspace) parseDependencies(r io.Reader) Subrepositories {
	var err error

	wdSbr := make([]Subrepository, 0, 100)
	latest := &Subrepository{wd: p.wd, branch: "master"}
	for err == nil {
		// I can do better than Fscanf
		var content bool
		content, err = scanSubrepository(latest, r)
		if content {
			wdSbr = append(wdSbr, latest.copy())
		}
	}
	if err != io.EOF {
		log.Fatalf("Error while reading .sbr: %s", err.Error())
	}
	return wdSbr
}

//scanSubrepository parse 'r' for a single subrepository definition.
// it updates 's' accordingly, and return content = true if at least path and remote have been read
// it is possible that it read path and remote (but no branch) get an EOF error and returns it.
func scanSubrepository(s *Subrepository, r io.Reader) (content bool, err error) {
	//now in the line scan for git path remote branch
	// branch beeing optional
	var kind string
	n, err := fmt.Fscanf(r, "%s %q %q %q\n", &kind, &s.rel, &s.remote, &s.branch)
	if n >= 3 {
		content = true
	}

	if n == 3 && err != io.EOF {
		err = nil //ignore those errors
		return
	}
	return
}
