package mrepo

import (
	"crypto/sha1"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

//projectRun is the result of a command execution on a given project
type projectRun struct {
	Name   string
	Cmd    string
	Args   []string
	Result string
}

func (p *projectRun) Head() string {
	return fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s \n", p.Name, p.Cmd, strings.Join(p.Args, " "))
}

// any function that can be started in a gorutine to flush the chan, and process results
type Outputer func(outputer <-chan projectRun)

func Default(outputer <-chan projectRun) {
	var count int
	for prjRun := range outputer {
		count++
		// default printing
		fmt.Print(prjRun.Head() + prjRun.Result)
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

func Cat(outputer <-chan projectRun) {
	for prjRun := range outputer {
		fmt.Print(prjRun.Result)
	}
}

//TODO
func Sum(outputer <-chan projectRun) {
	var total float64 = 0
	for prjRun := range outputer {
		cleaned := strings.Trim(prjRun.Result, " \n\r\t")
		res, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			total = math.NaN()
		} else {
			total += res
		}
	}
	fmt.Printf("Total: %v\n", total)
}

func Count(outputer <-chan projectRun) {
	hist := make(map[string]int)
	var count int
	for prjRun := range outputer {
		count++
		cleaned := strings.Trim(prjRun.Result, " \n\r\t")
		hist[cleaned]++
	}
	for k, v := range hist {
		fmt.Printf("\033[00;32m%v\033[00m : %s\n", v, k)

	}
	fmt.Printf("________________\n\033[00;32m%v\033[00m\n", count)
}

func Digest(outputer <-chan projectRun) {

	//we are going to sort prj by name first
	all := make([]projectRun, 0, 100)
	//first flush the outputer and store projects
	for prjRun := range outputer {
		all = append(all, prjRun)
	}
	sort.Sort(byName(all))

	h := sha1.New()
	for _, prjRun := range all {
		cleaned := strings.Trim(prjRun.Result, " \n\r\t")
		fmt.Fprint(h, cleaned)
	}
	fmt.Printf("digest: %x\n", h.Sum(nil))
}

//byName to sort any slice of projectRun byName !
type byName []projectRun

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
