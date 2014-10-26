package mrepo

import (
	"bufio"
	"io"
	"log"
)

type dependencyParser struct {
	wd   string // the working dir
	post DependencyProcessor
}

func (p *dependencyParser) ParseDependencies(r io.Reader) {
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

	p.post(dependencies)
}
