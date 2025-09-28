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
	s.router.GET("/admin/login", s.renderPage(s.ctx, pages.LoginPage))
	s.router.POST("/admin/login", s.renderPage(s.ctx, pages.LoginPage))
	// not really a frame, just redirects or throws error
	s.router.POST("/admin/logout", s.renderFrame(s.ctx, pages.LogoutFrame, RequireLogin()))

	s.router.GET("/admin", s.renderPage(s.ctx, pages.LinksPage, RequireLogin()))
	s.router.GET("/admin/links", s.renderPage(s.ctx, pages.LinksPage, RequireLogin()))
	s.router.GET("/admin/frame/links/new", s.renderFrame(s.ctx, pages.LinkNewFrame, RequireLogin()))
	s.router.GET("/admin/frame/links/update/:short_link", s.renderFrame(s.ctx, pages.LinkUpdateFrame, RequireLogin()))
	s.router.POST("/admin/links", s.renderFrame(s.ctx, pages.LinksFrame, RequireLogin()))
	s.router.PATCH("/admin/links/:short_link", s.renderFrame(s.ctx, pages.LinksFrame, RequireLogin()))
	s.router.DELETE("/admin/links/:short_link", s.renderFrame(s.ctx, pages.LinksFrame, RequireLogin()))
	s.router.GET("/admin/users", s.renderPage(s.ctx, pages.UsersPage, RequireLogin()))
	s.router.GET("/admin/frame/users/new", s.renderFrame(s.ctx, pages.UserNewFrame, RequireLogin()))
	s.router.GET("/admin/frame/users/update/:id", s.renderFrame(s.ctx, pages.UserUpdateFrame, RequireLogin()))
	s.router.POST("/admin/users", s.renderFrame(s.ctx, pages.UsersFrame, RequireLogin()))
	s.router.PATCH("/admin/users/:id", s.renderFrame(s.ctx, pages.UsersFrame, RequireLogin()))
	s.router.DELETE("/admin/users/:id", s.renderFrame(s.ctx, pages.UsersFrame, RequireLogin()))

	// short link handler as last route, catch all
	s.router.NotFound = handlerFuncAdapter(s.renderFrame(s.ctx, pages.ShortLinkHandler))
}

func (s *Server) setupGeneralRoutes() error {
	assets, err := s.ctx.Config.Assets()
	if err != nil {
		return errors.Wrap(err, "config.Assets")
	}

	s.router.ServeFiles("/dist/*filepath", assets.Dist)
	s.router.ServeFiles("/img/*filepath", assets.Img)

	return nil
}
