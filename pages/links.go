package pages

import (
	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

func LinksPage(req *foundation.Request) (*Page, error) {
	var body html.Blocks
	body.Add(html.Div(attr.Class("p-4 md:p-6 xl:p-12"),
		html.H2(attr.Class("text-2xl font-bold"),
			html.Text("Links"),
		),
	))

	// body.Add(html.P(attr.Class("text-center mt-4"),
	// 	html.Button(attr.Class("btn-outline").
	// 		DataAttr("controller", "toast-button").
	// 		DataAttr("action", "click->toast-button#toast"),
	// 		html.Text("Toast from front-end"),
	// 	),
	// ))
	// body.Add(html.Div(attr.Id("toaster").Class("toaster")))

	page := &Page{
		Title:   "Quick Links - Links",
		Body:    body,
		Sidebar: Sidebar{},
		Header:  Header{Title: "Links"},
	}

	return page, nil
}
