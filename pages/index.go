package pages

import (
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

func IndexPage(req *foundation.Request) (*Page, error) {
	log.Println("rendering index page")

	body := html.H1(attr.Class("text-3xl font-bold p-8 text-center"),
		html.Text(req.Config.Message),
	)
	page := &Page{
		Title: "Foundation",
		Body:  body,
	}

	return page, nil
}
