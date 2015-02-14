package cmd

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/ericaro/mrepo"
	"github.com/ericaro/mrepo/git"
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
	sort.Sort(byExecName(all))

	// now compute the sha1
	h := sha1.New()
	for _, x := range all {
		fmt.Fprint(h, x.Result)
	}
	fmt.Printf("%x\n", h.Sum(nil))
}

//byExecName to sort any slice of Execution by their Name !
type byExecName []Execution

func (a byExecName) Len() int           { return len(a) }
func (a byExecName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byExecName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type ExecCmd struct {
	cat, sum, count, digest *bool
	local                   *bool
}

func (c *ExecCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// output selection
	c.cat = fs.Bool("cat", false, "concatenate outputs, and print it")
	c.sum = fs.Bool("sum", false, "parse each output as a number and print out the total")
	c.count = fs.Bool("count", false, "count different outputs, and prints the resulting histogram")
	c.digest = fs.Bool("digest", false, "compute the sha1 digest of all outputs")
	c.local = fs.Bool("l", false, "start in the current working dir. Default is to start in the sbr workspace")

	return fs
}

func (c *ExecCmd) Run(args []string) {

	// use wd by default
	var wd string
	if *c.local {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			log.Fatalf("Error, cannot determine the current directory. %s\n", err.Error())
			os.Exit(CodeNoWorkingDir)
		}
	} else {
		wd = FindRootCmd()
	}
	//build the workspace, that is used to trigger all commands
	workspace := mrepo.NewWorkspace(wd)

	//again, passing the stdin, and stdout to the subprocess prevent: async, and ability to collect the outputs
	// for special outputers we need to collect outputs, so the 'special' var.
	// special => concurrent mode (because we need to collect outputs)
	// Therefore, selecting the output mode imply selecting "special"= true|false
	// and the ExecutionProcessor function
	xargs := make([]string, 0)
	if len(args) > 1 {
		xargs = args[1:]
	}
	name := args[0]
	executions := ExecConcurrently(workspace, name, xargs...)
	switch {
	case *c.cat:
		ExecutionCat(executions)
	case *c.sum:
		ExecutionSum(executions)
	case *c.count:
		ExecutionCount(executions)
	case *c.digest:
		ExecutionDigest(executions)
	default:
		ExecutionPrinter(executions)
	}
}

//ExecConcurently, for each `subrepository` in the working dir, execute the command `command` with arguments `args`.
// Each command is executed in non interactive mode (no access to stdin/stdout)
func ExecConcurrently(x *mrepo.Workspace, command string, args ...string) <-chan Execution {
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
			rel := x.Relativize(sub)
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", sub, command, strings.Join(args, " "))
			//executions <- head + string(out)
			result := string(out)
			result = strings.Trim(result, git.DefaultTrimCut)
			executions <- Execution{Name: sub, Rel: rel, Cmd: command, Args: args, Result: result}
		}(sub)
	}

	go func() {
		waiter.Wait()
		close(executions)
	}()
	return executions
}
