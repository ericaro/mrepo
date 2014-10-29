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
// Caveat, sync execution passes stdin/out to the subprocess that runs the command, so it can run in interactive mode,
// whereas async execution does not.
// async mode is required for statistical postprocessors.
type Workspace struct {
	wd             string       //current working dir
	mrepofilename  string       // the .mrepo filename (by default .mrepo)
	dependencyFile Dependencies // dependencies as declared in the .mrepo
	dependencies   Dependencies // dependencies as found in the workspace.
}

//NewWorkspace creates a new Workspace for a working dir.
func NewWorkspace(wd string) *Workspace {
	return &Workspace{
		wd:            wd,
		mrepofilename: ".mrepo",
	}
}
func (x *Workspace) Getwd() string {
	return x.wd
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

//start a gorutine in charge of scanning the local repo AND return the chan that will contain it.
func (x *Workspace) Scan() <-chan string {
	prjc := make(chan string)
	//the subrepo scanner function
	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if f.Name() == ".git" {
				// it's a repository file
				prjc <- filepath.Dir(path)
				//always skip the repository file
				return filepath.SkipDir
			}
		}
		return nil
	}

	go func() {
		defer close(prjc)
		filepath.Walk(x.wd, walker)
	}()
	return prjc
}

//ExecSync runs for each `subrepository` found by Scanner the  command `command` with arguments `args`
// It passes the stdin, stdout, and stderr to the subprocess. and wait for the result.
func (x *Workspace) ExecSync(command string, args ...string) {
	var count int

	for sub := range x.Scan() {
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

//Exec runs for each `subrepository` found by Scanner the  command `command` with arguments `args`.
// Each command is executed concurrently, and the outputs are collected (both err, and out).
func (x *Workspace) Exec(command string, args ...string) <-chan Execution {
	executions := make(chan Execution)
	var waiter sync.WaitGroup // to wait for all commands to return
	for sub := range x.Scan() {
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

//ExecQuery runs git queries for path, remote url, and branch on each subrepository, and then pushes the result for in a chan of Dependency
func (x *Workspace) DependencyWorkingDir() Dependencies {

	if x.dependencies == nil {

		dependencies := make(Dependencies, 0, 100)

		for prj := range x.Scan() {
			branch, err := GitBranch(prj)
			if err != nil {
				log.Fatalf("err getting branch %s", err.Error())
			}
			origin, err := GitRemoteOrigin(prj)
			if err != nil {
				log.Fatalf("err getting origin %s", err.Error())
			}
			rel := x.relpath(prj)
			if rel != "." {
				dependencies = append(dependencies, Dependency{
					wd:     x.wd,
					rel:    rel,
					remote: origin,
					branch: branch,
				})
			}
		}
		sort.Sort(dependencies)
		x.dependencies = dependencies
	}
	return x.dependencies
}

//DependencyFile returns a set of Dependencies, as declared in the .mrepo file
func (x *Workspace) DependencyFile() (dependencies Dependencies) {
	if x.dependencies == nil {
		file, err := os.Open(x.mrepofilename)
		if err == nil {
			defer file.Close()
			x.dependencyFile = x.parseDependencies(file) // for now, just parse
		} else {
			if os.IsNotExist(err) {
				fmt.Printf("dependency file %q does not exists. Skipping\n", x.mrepofilename)
			} else {
				fmt.Printf("Error reading dependency file %q: %s", x.mrepofilename, err.Error())
			}
		}
	}
	return x.dependencyFile
}

//ParseDependencies scans 'r' and fill []Dependency
func (p *Workspace) parseDependencies(r io.Reader) Dependencies {
	var err error

	dependencies := make([]Dependency, 0, 100)
	for err == nil {
		var kind string
		d := Dependency{wd: p.wd}
		// I can do better than Fscanf
		_, err = fmt.Fscanf(r, "%s %q %q %q\n", &kind, &d.rel, &d.remote, &d.branch)
		if err == nil {
			dependencies = append(dependencies, d)
		}
	}
	if err != io.EOF {
		log.Fatalf("Error while reading .mrepo: %s", err.Error())
	}
	return dependencies
}

func (x *Workspace) WriteDependencyFile(dependencies Dependencies) {
	f, err := os.OpenFile(x.mrepofilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("Cannot write dependency file: %s", err.Error())
		return
	}
	defer f.Close()
	sort.Sort(dependencies)
	dependencies.FormatMrepo(f)
}

//WorkingDirUpdates compare dependencies in the mrepo file with the one in the working dir.
// computes the changes to applied to the working dir, so that both are equals.
func (w *Workspace) WorkingDirUpdates() (ins, del Dependencies) {
	target := w.DependencyFile()
	current := w.DependencyWorkingDir()
	ins, del = current.Diff(target)
	return
}
