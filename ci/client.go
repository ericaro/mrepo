package ci

import (
	"errors"
	"log"

	"github.com/ericaro/sbr/format"
)

var (
	ErrUnreachable = errors.New("Cannot reach CI server")
)

type Client struct {
	*format.ProtoClient
}

func NewClient(server string) Client {
	return Client{format.NewClient(server)}
}

//turn a proto request into a response, error: dealing with protocol errors, and response error.
// return ErrUnreachable if this there was a communication error
// returns the protocol error otherwise
func (c Client) Request(req *format.Request) (resp *format.Response, err error) {
	resp, err = c.Proto(req)
	if err != nil {
		log.Printf("%v: %v", ErrUnreachable, err)
		return nil, ErrUnreachable
	}
	if resp.Error != nil {
		return resp, errors.New(resp.GetError())
	}
	return resp, nil
}

//AddJob add a job by it's name, remote and branch
func (c Client) AddJob(jobname, remote, branch string) (err error) {
	_, err = c.Request(
		&format.Request{
			Add: &format.AddRequest{
				Id: &format.Jobid{
					Name:   &jobname,
					Remote: &remote,
					Branch: &branch,
				},
			},
		})
	return
}

//Remove a job by its jobname
func (c Client) RemoveJob(jobname string) (err error) {
	_, err = c.Request(
		&format.Request{
			Remove: &format.RemoveRequest{
				Jobname: &jobname,
			},
		})
	return
}

//ListJobs return the list of all jobs.
//
// when 'refresh' (resp 'build') is true, the refresh (resp. build) log
// is sent too.
func (c Client) ListJobs(refresh, build bool) (jobs []*format.Job, err error) {

	resp, err := c.Request(
		&format.Request{
			List: &format.ListRequest{
				RefreshResult: &refresh,
				BuildResult:   &build,
			},
		})
	if err != nil {
		return nil, err
	}
	return resp.List.Jobs, nil
}

func (c Client) JobDetails(jobname string) (job *format.Job, err error) {
	resp, err := c.Request(
		&format.Request{
			Log: &format.LogRequest{Jobname: &jobname},
		})
	if err != nil {
		return nil, err
	}
	return resp.Log.Job, nil
}
