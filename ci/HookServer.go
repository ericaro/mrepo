package ci

import (
	"fmt"
	"net/http"
)

//HookServer is an http Server that implements:
// POST * : triggers a hearbeat
// GET  * : return a status code
type HookServer struct {
	Daemon
}

func NewHookServer(daemon Daemon) *HookServer { return &HookServer{daemon} }

//ServeHTTP defines the http server
func (s *HookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "POST":
		s.HeartBeats()

	case "GET":
		fmt.Fprint(w, s.Status())
	}
}
