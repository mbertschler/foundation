package pages

import (
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

func IndexPage(req *foundation.Request) (*Page, error) {
	log.Println("rendering index page")

	var body html.Blocks
	body.Add(html.H1(attr.Class("text-3xl font-bold p-8 text-center"),
		html.Text(req.Config.Message),
	))
	body.Add(html.P(attr.Class("text-center"),
		html.Button(attr.Class("btn-primary"), html.Text("Go!")),
	))
	page := &Page{
		Title: "Foundation",
		Body:  body,
	}

	return page, nil
}
