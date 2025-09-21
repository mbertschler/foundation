package pages

import (
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

func LinksPage(req *foundation.Request) (*Page, error) {
	var body html.Blocks
	body.Add(html.H1(attr.Class("text-3xl font-bold p-8 text-center"),
		html.Text(req.Config.Message),
	))

	body.Add(html.P(attr.Class("text-center mt-4"),
		html.Button(attr.Class("btn-outline").
			DataAttr("controller", "toast-button").
			DataAttr("action", "click->toast-button#toast"),
			html.Text("Toast from front-end"),
		),
	))
	body.Add(html.Div(attr.Id("toaster").Class("toaster")))

	page := &Page{
		Title:   "Quicklink - Links",
		Body:    body,
		Sidebar: Sidebar{},
		Header:  Header{Title: "Links"},
	}

	return page, nil
}
