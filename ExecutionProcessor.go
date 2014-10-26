package mrepo

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// ExecutionProcessor is a function that should process executions from the given chan
type ExecutionProcessor func(<-chan Execution)

//Execution is the result of a command Execution on a given project
// You get the project's name (the full path to the repository )
type Execution struct {
	Name   string
	Rel    string // relative path to the root
	Cmd    string
	Args   []string
	Result string
}

//Base returns the Base name for the project
func (e *Execution) Base() string {
	return filepath.Base(e.Name)

}

//DefaultPostProcessor just print a colored header and the result
func DefaultPostProcessor(source <-chan Execution) {
	var count int
	for x := range source {
		count++
		// default printing
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s \n%s\n", x.Rel, x.Cmd, strings.Join(x.Args, " "), x.Result)
	}

	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Cat ExecutionProcessor `cat` together all outputs.
func Cat(source <-chan Execution) {
	for x := range source {
		fmt.Print(x.Result)
	}
}

//Sum ExecutionProcessor try to parse the Execution output and sum it up.
// if it can parse it as a number it uses `NaN`.
func Sum(source <-chan Execution) {
	var total float64
	w := tabwriter.NewWriter(os.Stdout, 6, 8, 3, '\t', 0)

	for x := range source {
		cleaned := strings.Trim(x.Result, " \n\r\t")
		res, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			res = math.NaN()
		}
		fmt.Fprintf(w, "\t%v\t%s\n", res, x.Rel)
		total += res
	}
	w.Flush()
	fmt.Println("")
	fmt.Fprintf(w, "\t%v\t%s\n", total, "Total")
	w.Flush()

}

//Count ExecutionProcessor that count unique outputs
func Count(source <-chan Execution) {
	hist := make(map[string]int)
	var count int
	for x := range source {
		count++
		cleaned := strings.Trim(x.Result, " \n\r\t")
		hist[cleaned]++
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 3, '\t', 0)

	for k, v := range hist {
		fmt.Fprintf(w, "\033[00;32m%v\t\033[00m%s\n", v, k)
	}
	w.Flush()
	fmt.Println("")
	fmt.Fprintf(w, "\033[00;32m%v\t\033[00m%s\n", count, "Total")
	w.Flush()

}

//Digest ExecutionProcessor computes the digest of all outputs.
// Outputs are trimed of whitespaces. (` \n\r\t`)
func Digest(source <-chan Execution) {

	//we are going to sort prj by name first
	all := make([]Execution, 0, 100)
	//first flush the source and store projects
	for x := range source {
		all = append(all, x)
	}
	sort.Sort(byName(all))

	h := sha1.New()
	for _, x := range all {
		cleaned := strings.Trim(x.Result, " \n\r\t")
		fmt.Fprint(h, cleaned)
	}
	fmt.Printf("%x\n", h.Sum(nil))
}

//byName to sort any slice of Execution by their Name !
type byName []Execution

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
