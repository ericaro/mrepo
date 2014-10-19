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

//execution is the result of a command execution on a given project
// You get the project's name (the full path to the repository )
type execution struct {
	Name   string
	Cmd    string
	Args   []string
	Result string
}

//Base returns the Base name for the project
func (e *execution) Base() string {
	return filepath.Base(e.Name)

}

// PostProcessor is a function that should process executions from the given chan
type PostProcessor func(<-chan execution)

//Default PostProcessor: print a colored header and the result
func Default(source <-chan execution) {
	var count int
	for x := range source {
		count++
		// default printing
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s \n %s", x.Name, x.Cmd, strings.Join(x.Args, " "), x.Result)
	}

	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Cat PostProcessor `cat` together all outputs.
func Cat(source <-chan execution) {
	for x := range source {
		fmt.Print(x.Result)
	}
}

//Sum PostProcessor try to parse the execution output and sum it up.
// if it can parse it as a number it uses `NaN`.
func Sum(source <-chan execution) {
	var total float64 = 0
	w := tabwriter.NewWriter(os.Stdout, 6, 8, 3, '\t', 0)

	for x := range source {
		cleaned := strings.Trim(x.Result, " \n\r\t")
		res, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			res = math.NaN()
		}
		fmt.Fprintf(w, "\t%v\t%s\n", res, x.Base())
		total += res
	}
	w.Flush()
	fmt.Println("  __________")
	fmt.Fprintf(w, "\t%v\t%s\n", total, "")
	w.Flush()

}

//Count PostProcessor that count unique outputs
func Count(source <-chan execution) {
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
	fmt.Printf("________________\n")
	fmt.Fprintf(w, "\033[00;32m%v\t\033[00m%s\n", count, "Total")
	w.Flush()

}

//Digest PostProcessor computes the digest of all outputs.
// Outputs are trimed of whitespaces. (` \n\r\t`)
func Digest(source <-chan execution) {

	//we are going to sort prj by name first
	all := make([]execution, 0, 100)
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

//byName to sort any slice of execution by their Name !
type byName []execution

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
