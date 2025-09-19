package pages

import (
	"log"
	"net/http"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
	"github.com/pkg/errors"
)

func LoginPage(req *foundation.Request) (*Page, error) {
	page := &Page{
		Title: "Foundation - Login",
		Body:  loginFrame(nil),
	}
	return page, nil
}

func LoginFrame(req *foundation.Request) (html.Block, error) {
	var loginErr error
	switch req.Request.Method {
	case http.MethodPost:
		loginErr = postLogin(req)
		if loginErr == nil && req.User != nil {
			http.Redirect(req.Writer, req.Request, "/admin", http.StatusSeeOther)
			return nil, nil
		}
		if loginErr != nil {
			req.Writer.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
	return loginFrame(loginErr), nil
}

func postLogin(req *foundation.Request) error {
	err := auth.Login(req)
	if err != nil {
		log.Println("login error:", err)
		return errors.New("Invalid username or password.")
	}
	return nil
}

func loginFrame(err error) html.Block {
	var errBlock html.Block
	if err != nil {
		errBlock = html.Div(attr.Class("alert-destructive"),
			html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
				html.Elem("circle", attr.Attr("cx", "12").Attr("cy", "12").Attr("r", "10")),
				html.Elem("line", attr.Attr("x1", "12").Attr("x2", "12").Attr("y1", "8").Attr("y2", "12")),
				html.Elem("line", attr.Attr("x1", "12").Attr("x2", "12.01").Attr("y1", "16").Attr("y2", "16")),
			),
			html.H2(nil,
				html.Text("Login Error"),
			),
			html.Section(nil,
				html.Text(err.Error()),
			),
		)
	}

	return html.Div(attr.Id("login-frame").Class("min-h-screen grid place-items-center bg-gray-100"),
		html.Div(attr.Class("card max-w-md w-full"),
			html.Header(nil,
				html.H2(nil,
					html.Text("Login to Foundation"),
				),
				html.P(nil,
					html.Text("Enter your details below to login to your account."),
				),
				errBlock,
			),
			html.Section(nil,
				html.Form(attr.Id("login-form").Class("form grid gap-6").
					Method("POST").Action("/admin/login"),
					html.Div(attr.Class("grid gap-2"),
						html.Label(attr.For("login-form-username"),
							html.Text("Username"),
						),
						html.Input(attr.Type("text").Name("username").Id("login-form-username")),
					),
					html.Div(attr.Class("grid gap-2"),
						html.Label(attr.For("login-form-password"),
							html.Text("Password"),
						),
						html.Input(attr.Type("password").Name("password").Id("login-form-password")),
					),
				),
			),
			html.Footer(attr.Class("flex flex-col items-center gap-2"),
				html.Button(attr.Form("login-form").Type("submit").Class("btn w-full"),
					html.Text("Login"),
				),
				// html.Button(attr.Type("button").Class("btn-outline w-full"),
				// 	html.Text("Login with Google"),
				// ),
			),
		),
	)
}
