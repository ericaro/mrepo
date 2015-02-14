package ci

import (
	"errors"
	"log"

	"github.com/ericaro/mrepo/format"
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

func (c Client) AddJob(path, remote, branch string) (err error) {
	_, err = c.Request(
		&format.Request{
			Add: &format.AddRequest{
				Id: &format.Jobid{
					Name:   &path,
					Remote: &remote,
					Branch: &branch,
				},
			},
		})
	return
}

func (c Client) RemoveJob(path string) (err error) {
	_, err = c.Request(
		&format.Request{
			Remove: &format.RemoveRequest{
				Jobname: &path,
			},
		})
	return
}
