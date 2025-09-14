package pages

import (
	"fmt"
	"log"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
	"github.com/pkg/errors"
)

func IndexPage(req *foundation.Request) (*Page, error) {
	log.Println("rendering index page")

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

	allUsers, err := req.DB.Users.All(req.Context.Context)
	if err != nil {
		return nil, errors.Wrap(err, "All users")
	}

	body.Add(html.Div(attr.Class("p-4 md:p-6 xl:p-12"),
		html.H2(attr.Class("mx-auto max-w-screen-lg text-2xl font-bold pb-4"), html.Text("All Users")),
		html.Main(attr.Class("mx-auto relative flex w-full max-w-screen-lg gap-10"),

			usersTable(allUsers),
		),
	))

	page := &Page{
		Title: "Foundation",
		Body:  body,
	}

	return page, nil
}

func usersTable(users []*foundation.User) html.Block {
	var rows html.Blocks
	for _, u := range users {
		rows.Add(userTableRow(u))
	}

	return html.Div(attr.Class("overflow-x-auto w-full"),
		html.Table(attr.Class("table"),
			html.Caption(nil,
				html.Text("All known users."),
			),
			html.Thead(nil,
				html.Tr(nil,
					html.Th(nil,
						html.Text("ID"),
					),
					html.Th(nil,
						html.Text("Display Name"),
					),
					html.Th(nil,
						html.Text("User Name"),
					),
					html.Th(nil,
						html.Text("Created"),
					),
					html.Th(nil,
						html.Text("Updated"),
					),
				),
			),
			html.Tbody(nil,
				rows,
			),
		),
	)
}

func userTableRow(user *foundation.User) html.Block {
	return html.Tr(nil,
		html.Td(attr.Class("font-medium"),
			html.Text(fmt.Sprint(user.ID)),
		),
		html.Td(nil,
			html.Text(user.DisplayName),
		),
		html.Td(nil,
			html.Text(user.UserName),
		),
		html.Td(attr.Class("text-right"),
			html.Text(user.CreatedAt.Format("2006-01-02 15:04:05.000")),
		),
		html.Td(attr.Class("text-right"),
			html.Text(user.UpdatedAt.Format("2006-01-02 15:04:05.000")),
		),
	)
}
