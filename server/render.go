package server

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/pages"
	"github.com/mbertschler/html"
)

func (s *Server) renderPage(ctx *foundation.Context, fn pages.RenderFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		req := &foundation.Request{
			Context: ctx,
			Params:  params,
			Request: r,
		}

		page, err := fn(req)
		if err != nil {
			log.Println("RenderFunc error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = html.Render(w, page.RenderHTML())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
