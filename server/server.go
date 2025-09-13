package server

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/pages"
	"github.com/pkg/errors"
)

type Server struct {
	ctx    *foundation.Context
	router *httprouter.Router
}

func StartServer(ctx *foundation.Context) error {
	srv := &Server{
		ctx:    ctx,
		router: httprouter.New(),
	}

	srv.setupPageRoutes()

	err := srv.setupGeneralRoutes()
	if err != nil {
		return errors.Wrap(err, "setupGeneralRoutes")
	}

	return srv.start()
}

func (s *Server) setupPageRoutes() {
	s.router.GET("/", s.renderPage(s.ctx, pages.IndexPage))
}

func (s *Server) setupGeneralRoutes() error {
	assets, err := s.ctx.Config.Assets()
	if err != nil {
		return errors.Wrap(err, "config.Assets")
	}

	s.router.ServeFiles("/dist/*filepath", assets.Dist)
	s.router.ServeFiles("/img/*filepath", assets.Img)

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "page not found", http.StatusNotFound)
	})

	return nil
}

func (s *Server) start() error {
	hostPort := s.ctx.Config.HostPort
	log.Printf("starting server on http://%s", hostPort)
	go func() {
		http.ListenAndServe(hostPort, s.router)
	}()
	return nil
}
