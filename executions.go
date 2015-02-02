package mrepo

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

//this file contains function dealing with chan Execution (ie results of random command)

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

//ExecutionPrinter just print a colored header and the result
func ExecutionPrinter(source <-chan Execution) {
	var count int
	for x := range source {
		count++
		// default printing
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s \n%s\n", x.Rel, x.Cmd, strings.Join(x.Args, " "), x.Result)
	}

	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//ExecutionCat ExecutionProcessor `cat` together all outputs.
func ExecutionCat(source <-chan Execution) {
	for x := range source {
		fmt.Println(x.Result)
	}
}

//ExecutionSum  attempt to parse the Execution result as a number and sum it up.
// if it can parse it as a number it uses `NaN`.
func ExecutionSum(source <-chan Execution) {
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

//ExecutionCount counts different outputs
func ExecutionCount(source <-chan Execution) {
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

//ExecutionDigest computes the digest of all execution results concatenated.
// Outputs are trimed of whitespaces. (` \n\r\t`)
func ExecutionDigest(source <-chan Execution) {

	//we are going to sort prj by name first
	all := make([]Execution, 0, 100)

	//first flush the source, store them, and sort them
	// because we need to ompute the digest in a repetitive order
	for x := range source {
		all = append(all, x)
	}
	sort.Sort(byName(all))

	// now compute the sha1
	h := sha1.New()
	for _, x := range all {
		fmt.Fprint(h, x.Result)
	}
	fmt.Printf("%x\n", h.Sum(nil))
}

//byName to sort any slice of Execution by their Name !
type byName []Execution

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
