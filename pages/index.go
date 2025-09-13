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

	clickHandler := `document.dispatchEvent(new CustomEvent('basecoat:toast', {
	    detail: {
	      config: {
	        category: 'success',
	        title: 'Success',
	        description: 'A success toast called from the front-end.',
	        cancel: {
	          label: 'Dismiss'
	        }
	      }
	    }
	  }))
	`
	body.Add(html.P(attr.Class("text-center mt-4"),
		html.Button(attr.Class("btn-outline").Attr("onclick", clickHandler),
			html.Text("Toast from front-end"),
		),
	))
	body.Add(html.Div(attr.Id("toaster").Class("toaster")))

	page := &Page{
		Title: "Foundation",
		Body:  body,
	}

	return page, nil
}
