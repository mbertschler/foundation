package server

import (
	"log"
	"net/http"

	"github.com/mbertschler/foundation"
)

type Server struct {
	ctx     *foundation.Context
	handler *http.ServeMux
}

func StartServer(ctx *foundation.Context) error {
	srv := &Server{
		ctx:     ctx,
		handler: http.NewServeMux(),
	}

	srv.handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(ctx.Config.Message))
	})

	return srv.start()
}

func (s *Server) start() error {
	hostPort := s.ctx.Config.HostPort
	log.Printf("starting server on http://%s", hostPort)
	go func() {
		http.ListenAndServe(hostPort, s.handler)
	}()
	return nil
}
