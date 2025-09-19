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

func RunServer(ctx *foundation.Context) error {
	srv := &Server{
		ctx:    ctx,
		router: httprouter.New(),
	}

	srv.setupPageRoutes()

	err := srv.setupGeneralRoutes()
	if err != nil {
		return errors.Wrap(err, "setupGeneralRoutes")
	}

	hostPort := srv.ctx.Config.HostPort
	log.Printf("starting server on http://%s", hostPort)
	return http.ListenAndServe(hostPort, srv.router)
}

func (s *Server) setupPageRoutes() {
	s.router.Handler("GET", "/", http.RedirectHandler("/admin", http.StatusFound))
	s.router.GET("/admin", s.renderPage(s.ctx, pages.IndexPage))
	s.router.GET("/admin/login", s.renderPage(s.ctx, pages.LoginPage))
	s.router.POST("/admin/login", s.renderFrame(s.ctx, pages.LoginFrame))

	s.router.GET("/admin/frame/users/new", s.renderFrame(s.ctx, pages.UserNewFrame))
	s.router.GET("/admin/frame/users/update/:id", s.renderFrame(s.ctx, pages.UserUpdateFrame))
	s.router.POST("/admin/users", s.renderFrame(s.ctx, pages.UsersFrame))
	s.router.PATCH("/admin/users/:id", s.renderFrame(s.ctx, pages.UsersFrame))
	s.router.DELETE("/admin/users/:id", s.renderFrame(s.ctx, pages.UsersFrame))
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
