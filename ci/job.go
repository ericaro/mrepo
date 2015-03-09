package ci

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ericaro/mrepo/format"
	"github.com/ericaro/mrepo/git"
	"github.com/ericaro/mrepo/makefile"
	"github.com/ericaro/mrepo/sbr"
)

//job is the main object in a ci. it represent a project to be build.
// it is configured by a unique name, a remote url (git url to checkout the project)
// and a branch to checkout.
type job struct {
	name   string
	remote string
	branch string
	// cmd      string    // the command executed as a CI (default `make`)
	// args     []string  // args of the ci command default `ci`

	//other fields are local one.
	at       *time.Timer // timer to deal with the delay
	refresh  execution   // info about the refresh execution
	build    execution   // info about the build execution
	execLock sync.Mutex  // there are still issues with this lock (in relation with delete)
}

func RunJobNow(name, remote, branch string) {
	j := &job{name: name, remote: remote, branch: branch}
	j.doRun()
	fmt.Println(j.refresh.result.String())
	fmt.Println(j.build.result.String())
	fmt.Println("done")

}

//Marshal serialize all information into a format.Job object.
func (j *job) Marshal() *format.Job { return j.Status(true, true) }

func (j *job) State() Status {

	if j.refresh.IsRunning() || j.build.IsRunning() {
		return StatusRunning
	}

	if j.refresh.errcode != 0 || j.build.errcode != 0 {
		return StatusKO
	}

	return StatusOK
}

//Status serialize information into a format.Job object.
//
// if withRrefresh, it will include the refresh output
//
// if withBuild if will include the build output
func (j *job) Status(withRefresh, withBuild bool) *format.Job {
	return &format.Job{
		Id: &format.Jobid{
			Name:   &j.name,
			Remote: &j.remote,
			Branch: &j.branch,
		},
		Refresh: j.refresh.Status(withRefresh),
		Build:   j.build.Status(withBuild),
	}
}

//Unmarshal initialise the current job with values from the format.Job message
func (j *job) Unmarshal(f *format.Job) error {

	id := f.Id
	j.name = id.GetName()
	j.remote = id.GetRemote()
	j.branch = id.GetBranch()

	if err := j.refresh.Unmarshal(f.GetRefresh()); err != nil {
		return err
	}
	if err := j.build.Unmarshal(f.GetBuild()); err != nil {
		return err
	}
	return nil
}

//Run waits for a while and then start a refresh/build execution.
func (j *job) Run() {

	j.RunWithDelay(10 * time.Second)
}

func (j *job) RunWithDelay(delay time.Duration) {
	// nobody can "change" at right now
	if j.at == nil { // first time
		j.at = time.AfterFunc(delay, j.doRun) // schedule
	} else {
		delayed := j.at.Reset(delay) //reset is enough
		//the bool is to tell if the action has been postponed or simply rescheduled
		if delayed {
			log.Printf("job[%q].execution.redelayed:%v", j.name, delay)
		}
	}
	log.Printf("job[%q].execution.delayed:%v", j.name, delay)
}

//doRun really execute the run
func (j *job) doRun() {
	log.Printf("job[%q].pull", j.name)
	j.Refresh()
	log.Printf("job[%q].build", j.name)
	j.Build()
}

//Refresh the current job.
//
// Skip if there is an ongoing job.
func (j *job) Refresh() {
	j.execLock.Lock()
	defer j.execLock.Unlock()

	//to start we refresh all information: buffer, and start time.
	j.refresh.result = new(bytes.Buffer)
	j.refresh.start = time.Now() // mark the job as started
	j.refresh.errcode = 0        // no semantic here... yet
	defer func() {               // we will update stuff at the end
		j.refresh.end = time.Now() // mark the job as ended at the end of this call.
	}()
	// do the job now and return
	if err := j.dorefresh(j.refresh.result); err != nil {
		j.refresh.errcode = -1 // no semantic here... yet
		fmt.Fprintln(j.refresh.result, err.Error())
	} else {
		j.refresh.errcode = 0
	}
	log.Printf("job[%q].pull.done", j.name)
}

//Build the current job
func (j *job) Build() {

	j.execLock.Lock()
	defer j.execLock.Unlock()

	// check that the version has changed
	/* temp deactivated  */
	if j.build.version == j.refresh.version {
		// currently uptodate, nothing to do
		log.Printf("job[%q].build.skip", j.name)
		return
	}
	/**/

	// I'm gonna run
	// I'm under the protection of the lock
	// mark the version has built
	j.build.result = new(bytes.Buffer)
	j.build.start = time.Now() // mark the job as started
	defer func() {
		j.build.version = j.refresh.version
		j.build.end = time.Now() // mark the job as ended at the end of this call.
	}()

	// do the job now and return
	if err := j.dobuild(j.build.result); err != nil {
		j.build.errcode = -1 // no semantic here... yet
		fmt.Fprintln(j.build.result, err.Error())
	} else {
		j.build.errcode = 0
	}
	log.Printf("job[%q].build.done", j.name)
}

func (j *job) dobuild(w io.Writer) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "working dir: %s\n", wd)

	fmt.Fprintf(w, "%s $ make ci\n", filepath.Join(wd, j.name))
	return makefile.Run(filepath.Join(wd, j.name), "ci", w)
}

//dorefresh actually run the refresh command, it is unsafe to call it without caution. It should only update errcode, and result
func (j *job) dorefresh(w io.Writer) error {

	// NB: this is  a mammoth function. I know it, but I wasn't able to
	// split it down before having done everything.
	//
	// Step by step I will extract subffunctions to appropriate set of objects
	//

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = os.Stat(j.name)
	if os.IsNotExist(err) { // target does not exist, make it.
		fmt.Fprintf(w, "%s dir does not exists. Will create one.\n", j.name)
		result, err := git.Clone(wd, j.name, j.remote, j.branch)
		fmt.Fprintln(w, result)
		if err != nil {
			return err
		}
	}

	wk := sbr.NewWorkspace(filepath.Join(wd, j.name))
	ch := sbr.NewCheckouter(wk, w)
	ch.SetFastForwardOnly(true)
	ch.SetPrune(true)

	digest, err := ch.Checkout()
	if err != nil {
		return err
	}
	copy(j.refresh.version[:], digest)
	return nil
}
