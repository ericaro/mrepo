package mrepo

import (
	"bufio"
	"io"
	"log"
)

type dependencyParser struct {
	wd string // the working dir
}

func (p *dependencyParser) ParseDependencies(r io.Reader) <-chan Dependency {
	dependencies := make(chan Dependency)
	go func() {

		scanner := bufio.NewScanner(r)
		//use a word splitter
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			rel := scanner.Text()
			if !scanner.Scan() {
				log.Fatalf("missing remote definition.")
			}
			remote := scanner.Text()
			if !scanner.Scan() {
				log.Fatalf("missing branch definition.")
			}
			branch := scanner.Text()

			//log.Printf("scanned to: git clone %s -b %s %s", remote, branch, rel)
			dependencies <- Dependency{
				rel:    rel,
				remote: remote,
				branch: branch,
				wd:     p.wd,
			}

		}
		close(dependencies) //done parsing
	}()
	return dependencies
}

//MergeDependencies reads target chan of dependency and current one, an generates two chan
// one for the insertion to be made to current to be equal to target
// one for the deletion to be made to current to be equal to target
//later, maybe we'll add update for branches
func MergeDependencies(target, current <-chan Dependency) (insertion, deletion <-chan Dependency) {
	targets := make(map[string]Dependency, 100)
	currents := make(map[string]Dependency, 100)

	ins, del := make(chan Dependency), make(chan Dependency)
	go func() {

		//first flush the targets and currents
		for x := range target {
			targets[x.rel] = x
		}
		for x := range current {
			currents[x.rel] = x
		}

		//then compute the diffs

		for id, t := range targets { // for each target
			_, exists := currents[id]
			if !exists { // if missing , create an insert
				ins <- t
			}
		}
		close(ins)

		for id, c := range currents { // for each current
			_, exists := targets[id]
			if !exists { // locally exists, but not in target, it's a deletion
				del <- c
			}
		}
		close(del)
	}()
	return ins, del
}
