// package ci handle ci job execution, persistence, andwsebservice.
package ci

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ericaro/mrepo/format"
	"github.com/golang/protobuf/proto"
)

type Status int

const (
	StatusKO      = 0
	StatusRunning = 1
	StatusOK      = 2
)

//Daemon defines the API for a Continuous Integration Server.
type Daemon interface {
	// Heartbeats notifies the daemon to start all jobs
	HeartBeats()
	// the AND of all statuses
	Status() Status
	AddJob(path, remote, branch string) error
	RemoveJob(path string) error
	ListJobs(refreshResult, buildResult bool) *format.ListResponse
	JobDetails(job string) *format.LogResponse
	// marshal the internal configuration into this protobuf message
	Marshal() *format.Server
	// Unmarshal from this protobuf message
	Unmarshal(*format.Server) error
}

//NewDaemon creates a new instance given a working dir, and a dbfile.
//
// dbfile must be a file containing a *format.Server message serialized message.
// if it does not exists, the daemon will start from scratch.
//
// when the daemon receives a interrupt, kill or SIGTERM signal it persists
// its internal state into the dbfile.
func NewDaemon(wd, dbfile string) (daemon Daemon, err error) {

	//Creates the daemon
	daemon = &ci{wd: wd, jobs: make(map[string]*job)}

	// read from disk if needed
	_, err = os.Stat(dbfile)

	if err == nil || !os.IsNotExist(err) {
		log.Printf("daemon.loading:%q", dbfile)
		// file exists, read the file
		file, err := os.Open(dbfile)
		if err != nil {
			log.Printf("daemon.opening.error:%q", err.Error())
			return daemon, err
		}
		defer file.Close()

		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("daemon.reading.error:%q", err.Error())
			return daemon, err
		}

		// now protobuf read the content
		f := new(format.Server)
		err = proto.Unmarshal(b, f)
		if err != nil {
			log.Printf("daemon.parsing.error:%q", err.Error())
			return daemon, err
		}

		// the format is read, init the server
		err = daemon.Unmarshal(f)
		if err != nil {
			log.Printf("daemon.unmarshalling.error:%q", err.Error())
			return daemon, err
		}
	}
	// now the ci is fully created or unmarshaled
	//just log the job found
	for i, n := range daemon.ListJobs(false, false).GetJobs() {
		log.Printf("    daemon.job[%v]:%q,\n", i, n.GetId().GetName())
	}

	// register a syscall hook to persist it on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for s := range c {
			log.Printf("daemon.interruption:%q", s.String())
			//marshalling
			b, err := proto.Marshal(daemon.Marshal())
			if err != nil {
				log.Printf("daemon.marshal.error:%q", err.Error())
				os.Exit(-1)
			}

			err = ioutil.WriteFile(filepath.Join(wd, dbfile), b, os.ModePerm)
			if err != nil {
				log.Printf("daemon.marshal.write.error:%q", err.Error())
				os.Exit(-1)
			}

			log.Printf("daemon.persisted")
			os.Exit(0)
		}
	}()

	log.Printf("daemon.ready")
	return daemon, nil
}

var _ Daemon = (*ci)(nil)

// ci is a collection of jobs. It implements Daemon
type ci struct {
	jobs       map[string]*job // path -> job
	wd         string          // absolute path to the working dir
	heartbeats int
}

//JobDetails returns all the details for a job.
func (c *ci) JobDetails(jobname string) *format.LogResponse {
	j := c.jobs[jobname]
	return &format.LogResponse{
		Job: j.Status(true, true),
	}
}

//Status return the global status.
// its the maximum status OK, KO, Running.
func (c *ci) Status() (status Status) {
	//state is all about sorting by "priority", and taking the max.
	// priority is not the natural order for Status (that is an int though)
	// because the zero value for a status has to be KO
	//wherease in a Daemon status the zero value is Ok (no job, no ko !)

	//we map status to a priority
	priority := map[Status]int{
		StatusOK:      0,
		StatusKO:      1,
		StatusRunning: 2,
	}

	current := 0      // current priority level
	status = StatusOK // the lowest priority also the value when empty

	for _, j := range c.jobs {
		js := j.State()
		p := priority[js]
		if p > current {
			current, status = p, js
		}
	}
	return
}

//ListJobs return a format.ListResponse describing all jobs.
// refreshResult = true means to add the output of the refresh action.
func (c *ci) ListJobs(refreshResult, buildResult bool) *format.ListResponse {

	//make and fill the slice
	js := make([]*format.Job, 0, len(c.jobs))
	for _, j := range c.jobs {
		js = append(js, j.Status(refreshResult, buildResult))
	}

	return &format.ListResponse{
		Jobs: js,
	}
}

//HeartBeat count incoming commits, and schedule a build
func (c *ci) HeartBeats() {
	c.heartbeats++
	log.Printf("daemon.heartbeat:%v", c.heartbeats)
	for _, j := range c.jobs {
		j.Run() // I don't need to fork here, because Run() already handles that.
	}
}

//AddJob add a new job
func (c *ci) AddJob(path, remote, branch string) error {
	if _, exists := c.jobs[path]; exists {
		return fmt.Errorf("a job named %q already exists.", path)
	}
	c.jobs[path] = &job{name: path,
		remote: remote,
		branch: branch,
	}
	return nil
}

//RemoveJob remove from the daemon, and prune the it's working dir.
func (c *ci) RemoveJob(path string) error {
	//TODO(EA) I've got a bug here: hard to reproduce:
	// add a job and remove it shortly after (the hard part is "how much shortly")
	if _, exists := c.jobs[path]; exists {
		//remove from the daemon server
		delete(c.jobs, path)

		//remove from local filesystem
		if err := os.RemoveAll(path); err != nil {
			if os.IsNotExist(err) {
				return nil //ok
			} else {
				return fmt.Errorf("cannot remove %q working directory: %v", path, err)
			}
		}
	}
	return nil
}

// the main feature for a ci is to edit jobs, and persist them.

//Marshal export internal state into format.Server protobuf message
func (c *ci) Marshal() *format.Server {
	jobs := make([]*format.Job, 0, 100)
	for _, j := range c.jobs {
		jobs = append(jobs, j.Marshal())
	}
	return &format.Server{Jobs: jobs}
}

//Unmarshal replace internal content with protobuf message
func (c *ci) Unmarshal(f *format.Server) error {

	// clean up the current object
	c.jobs = make(map[string]*job)

	for _, j := range f.Jobs {

		jb := job{}
		jb.Unmarshal(j)
		c.jobs[jb.name] = &jb

	}
	return nil
}
