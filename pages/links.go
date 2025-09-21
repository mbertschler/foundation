package pages

import (
	"fmt"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
	"github.com/pkg/errors"
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

	var userText string
	if req.User != nil {
		userText = fmt.Sprintf("User: ID %d Username %q Name %q",
			req.User.ID, req.User.UserName, req.User.DisplayName)
	} else {
		userText = "User: not logged in"
	}

	body.Add(html.Div(attr.Class("p-4 md:p-6 xl:p-12"),
		html.H3(attr.Class("text-xl font-bold mt-8 mb-4"), html.Text("Session and User Info")),
		html.Pre(attr.Class("bg-gray-100 p-4 rounded whitespace-pre-wrap"),
			html.Text(fmt.Sprintf("Session: ID %q UserID %t %d\n",
				req.Session.ID, req.Session.UserID.Valid, req.Session.UserID.Int64)),
			html.Text(userText),
		),
		html.Form(attr.Class("mt-8").Method("POST").Action("/admin/logout"),
			html.Button(attr.Class("btn-primary").Type("submit"),
				html.Text("Logout"),
			),
		),
	))

	// Add user management frame

	usersFrame, err := UsersFrame(req)
	if err != nil {
		return nil, errors.Wrap(err, "usersFrame")
	}
	body.Add(usersFrame)

	page := &Page{
		Title:   "Quicklink - Links",
		Body:    body,
		Sidebar: Sidebar{},
		Header:  Header{Title: "Links"},
	}

	return page, nil
}
