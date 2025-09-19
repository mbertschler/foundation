package server

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/foundation/pages"
	"github.com/mbertschler/html"
)

func (s *Server) renderPage(ctx *foundation.Context, fn pages.PageFunc) httprouter.Handle {
	return s.renderFrame(ctx, func(req *foundation.Request) (html.Block, error) {
		page, err := fn(req)
		if err != nil {
			return nil, err
		}
		return page.RenderHTML(), nil
	})
}

func (s *Server) renderFrame(ctx *foundation.Context, fn pages.FrameFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		req := &foundation.Request{
			Context: ctx,
			Writer:  w,
			Request: r,
			Params:  params,
		}

		sess, err := auth.GetOrCreateSession(req)
		if err != nil {
			log.Println("GetOrCreateSession error:", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		req.Session = sess

		if sess.UserID.Valid {
			user, err := req.DB.Users.ByID(req.Context, sess.UserID.Int64)
			if err != nil {
				log.Println("Users.ByID error:", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			req.User = user
		}

		block, err := fn(req)
		if err != nil {
			log.Println("Render error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = html.Render(w, block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
