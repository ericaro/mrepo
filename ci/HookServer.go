package ci

import (
	"fmt"
	"net/http"
)

type HookServer struct {
	daemon Daemon
}

func NewHookServer(daemon Daemon) *HookServer { return &HookServer{daemon} }

//ServeHTTP defines the http server
func (s *HookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// any incoming request here is a valid hook, and triggers a refresh / build
		s.daemon.HeartBeats()
	}
	if r.Method == "GET" {
		status := fmt.Sprintf("%v", s.daemon.Status())
		w.Write(([]byte)(status))

	}

}
