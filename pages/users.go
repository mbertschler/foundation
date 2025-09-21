package pages

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/foundation/pages/components"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
	"github.com/pkg/errors"
)

func UsersPage(req *foundation.Request) (*Page, error) {
	var body html.Blocks
	body.Add(html.P(attr.Class(" mt-4"),
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
		Title:   "Quicklink - Users",
		Sidebar: Sidebar{},
		Header: Header{
			Title: "Users",
		},
		Body: body,
	}

	return page, nil
}

func UsersFrame(req *foundation.Request) (html.Block, error) {
	switch req.Request.Method {
	case http.MethodPost:
		err := postNewUser(req)
		if err != nil {
			return nil, errors.Wrap(err, "postNewUser")
		}
	case http.MethodPatch:
		err := patchUser(req)
		if err != nil {
			return nil, errors.Wrap(err, "patchUser")
		}
	case http.MethodDelete:
		log.Println("DELETE user request")
		err := deleteUser(req)
		if err != nil {
			return nil, errors.Wrap(err, "deleteUser")
		}
	}

	allUsers, err := req.DB.Users.All(req.Context.Context)
	if err != nil {
		return nil, errors.Wrap(err, "All users")
	}

	return html.Elem("turbo-frame", attr.Id("users-frame"),
		html.Div(attr.Class("p-4 md:p-6 xl:p-12"),
			html.Main(attr.Class("mx-auto relative w-full max-w-screen-lg gap-10"),
				html.H2(attr.Class("mx-auto max-w-screen-lg text-2xl font-bold pb-4"), html.Text("All Users")),
				usersTable(allUsers),
				newUserForm(),
			),
		),
		html.Elem("turbo-frame", attr.Id("user-dialog-frame")),
	), nil
}

func postNewUser(req *foundation.Request) error {
	r := req.Request
	err := r.ParseForm()
	if err != nil {
		http.Error(req.Writer, "Failed to parse form", http.StatusBadRequest)
		return errors.Wrap(err, "ParseForm")
	}

	displayName := r.FormValue("display_name")
	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("Received new user data:")
	log.Printf("Display Name: %s", displayName)
	log.Printf("Username: %s", username)
	log.Printf("Password: %s", password)

	exists, err := req.DB.Users.ExistsByUsername(req.Context.Context, username)
	if err != nil {
		return errors.Wrap(err, "ExistsByUsername")
	}
	if exists {
		http.Error(req.Writer, fmt.Sprintf("Username %q already exists", username), http.StatusConflict)
		return errors.New("username exists")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return errors.Wrap(err, "HashPassword")
	}

	user := &foundation.User{
		DisplayName:    displayName,
		UserName:       username,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err = req.DB.Users.Insert(req.Context, user)
	if err != nil {
		return errors.Wrap(err, "Insert user")
	}

	log.Printf("Inserted new user with ID %d", user.ID)
	return nil
}

func patchUser(req *foundation.Request) error {
	r := req.Request
	err := r.ParseForm()
	if err != nil {
		http.Error(req.Writer, "Failed to parse form", http.StatusBadRequest)
		return errors.Wrap(err, "ParseForm")
	}

	// Extract user ID from URL path
	userIDStr := req.Params.ByName("id")
	if userIDStr == "" {
		http.Error(req.Writer, "User ID is required", http.StatusBadRequest)
		return errors.New("missing user ID")
	}

	var userID int64
	_, err = fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		http.Error(req.Writer, "Invalid user ID", http.StatusBadRequest)
		return errors.Wrap(err, "invalid user ID")
	}

	// Get existing user
	existingUser, err := req.DB.Users.ByID(req.Context.Context, userID)
	if err != nil {
		http.Error(req.Writer, "User not found", http.StatusNotFound)
		return errors.Wrap(err, "user not found")
	}

	displayName := r.FormValue("display_name")
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check if username is being changed and if it already exists
	if username != existingUser.UserName {
		exists, err := req.DB.Users.ExistsByUsername(req.Context.Context, username)
		if err != nil {
			return errors.Wrap(err, "ExistsByUsername")
		}
		if exists {
			http.Error(req.Writer, fmt.Sprintf("Username %q already exists", username), http.StatusConflict)
			return errors.New("username exists")
		}
	}

	// Update user fields
	existingUser.DisplayName = displayName
	existingUser.UserName = username
	if password != "" {
		hashedPassword, err := auth.HashPassword(password)
		if err != nil {
			return errors.Wrap(err, "HashPassword")
		}
		existingUser.HashedPassword = hashedPassword
	}
	existingUser.UpdatedAt = time.Now()

	err = req.DB.Users.Update(req.Context.Context, existingUser)
	if err != nil {
		return errors.Wrap(err, "Update user")
	}

	log.Printf("Updated user with ID %d", userID)
	return nil
}

func deleteUser(req *foundation.Request) error {
	// Extract user ID from URL path
	userIDStr := req.Params.ByName("id")
	if userIDStr == "" {
		http.Error(req.Writer, "User ID is required", http.StatusBadRequest)
		return errors.New("missing user ID")
	}

	var userID int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		http.Error(req.Writer, "Invalid user ID", http.StatusBadRequest)
		return errors.Wrap(err, "invalid user ID")
	}

	// Check if user exists
	_, err = req.DB.Users.ByID(req.Context.Context, userID)
	if err != nil {
		http.Error(req.Writer, "User not found", http.StatusNotFound)
		return errors.Wrap(err, "user not found")
	}

	// Delete user
	err = req.DB.Users.Delete(req.Context.Context, userID)
	if err != nil {
		return errors.Wrap(err, "Delete user")
	}

	log.Printf("Deleted user with ID %d", userID)
	return nil
}

func usersTable(users []*foundation.User) html.Block {
	var rows html.Blocks
	for _, u := range users {
		rows.Add(userTableRow(u))
	}

	return html.Div(attr.Class("overflow-x-auto w-full"),
		html.Table(attr.Class("table"),
			html.Caption(nil,
				html.Text("All known users. Rendered at "+fmt.Sprint(time.Now().Format("15:04:05.000"))),
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
					html.Th(nil,
						html.Text("Actions"),
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
		html.Td(nil,
			html.A(attr.Href(fmt.Sprintf("/admin/frame/users/update/%d", user.ID)).Class("btn-ghost").Attr("data-turbo-frame", "user-dialog-frame"),
				html.Text("Edit"),
			),
		),
	)
}

func newUserForm() html.Block {
	return html.Blocks{
		html.A(attr.Href("/admin/frame/users/new").Class("btn-outline").Attr("data-turbo-frame", "user-dialog-frame"),
			html.Text("Add New User"),
		),
	}
}

func UserNewFrame(req *foundation.Request) (html.Block, error) {
	return html.Elem("turbo-frame", attr.Id("user-dialog-frame"),
		html.Dialog(attr.Id("new-user-dialog").Class("dialog w-full sm:max-w-[425px] max-h-[612px]").Attr("aria-labelledby", "new-user-dialog-title").Attr("aria-describedby", "new-user-dialog-description").Attr("onclick", "if (event.target === this) this.close()"),
			html.Article(nil,
				html.Header(nil,
					html.H2(attr.Id("new-user-dialog-title"),
						html.Text("Add New User"),
					),
					html.P(attr.Id("new-user-dialog-description"),
						html.Text("Enter the details for the new user."),
					),
				),
				html.Section(nil,
					html.Form(attr.Method("POST").Action("/admin/users").Class("form grid gap-4").Attr("data-turbo-frame", "users-frame"),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For("display-name"),
								html.Text("Display Name"),
							),
							html.Input(attr.Type("text").Name("display_name").Id("display-name").Required("").Autofocus("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For("username"),
								html.Text("Username"),
							),
							html.Input(attr.Type("text").Name("username").Id("username").Required("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For("password"),
								html.Text("Password"),
							),
							html.Input(attr.Type("password").Name("password").Id("password").Required("")),
						),
						html.Div(attr.Class("flex justify-end gap-2 mt-4"),
							html.Button(attr.Type("button").Class("btn-outline").Attr("onclick", "this.closest('dialog').close()"),
								html.Text("Cancel"),
							),
							html.Button(attr.Type("submit").Class("btn"),
								html.Text("Create User"),
							),
						),
					),
				),
				html.Button(attr.Type("button").Attr("aria-label", "Close dialog").Attr("onclick", "this.closest('dialog').close()"),
					html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").Class("lucide lucide-x-icon lucide-x"),
						html.Elem("path", attr.Attr("d", "M18 6 6 18")),
						html.Elem("path", attr.Attr("d", "m6 6 12 12")),
					),
				),
			),
		),
		html.Script(nil, html.JS("document.getElementById('new-user-dialog').showModal();")),
	), nil
}

func UserUpdateFrame(req *foundation.Request) (html.Block, error) {
	userIDStr := req.Params.ByName("id")
	if userIDStr == "" {
		return nil, errors.New("user ID is required")
	}

	var userID int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user ID")
	}

	user, err := req.DB.Users.ByID(req.Context.Context, userID)
	if err != nil {
		return nil, errors.Wrap(err, "user not found")
	}

	return html.Elem("turbo-frame", attr.Id("user-dialog-frame"),
		html.Dialog(attr.Id(fmt.Sprintf("edit-user-dialog-%d", user.ID)).Class("dialog w-full sm:max-w-[425px] max-h-[612px]").Attr("aria-labelledby", fmt.Sprintf("edit-user-dialog-title-%d", user.ID)).Attr("aria-describedby", fmt.Sprintf("edit-user-dialog-description-%d", user.ID)).Attr("onclick", "if (event.target === this) this.close()"),
			html.Article(nil,
				html.Header(nil,
					html.H2(attr.Id(fmt.Sprintf("edit-user-dialog-title-%d", user.ID)),
						html.Text("Edit User"),
					),
					html.P(attr.Id(fmt.Sprintf("edit-user-dialog-description-%d", user.ID)),
						html.Text("Make changes to the user details."),
					),
				),
				html.Section(nil,
					html.Form(attr.Id("delete-user").Method("DELETE").Action(fmt.Sprintf("/admin/users/%d", user.ID)).Attr("data-turbo-frame", "users-frame")), //.Attr("onsubmit", "this.closest('dialog').close(); return true;"),
					html.Form(attr.Method("PATCH").Action(fmt.Sprintf("/admin/users/%d", user.ID)).Class("form grid gap-4").Attr("data-turbo-frame", "users-frame"),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For(fmt.Sprintf("edit-display-name-%d", user.ID)),
								html.Text("Display Name"),
							),
							html.Input(attr.Type("text").Name("display_name").Id(fmt.Sprintf("edit-display-name-%d", user.ID)).Value(user.DisplayName).Required("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For(fmt.Sprintf("edit-username-%d", user.ID)),
								html.Text("Username"),
							),
							html.Input(attr.Type("text").Name("username").Id(fmt.Sprintf("edit-username-%d", user.ID)).Value(user.UserName).Required("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For(fmt.Sprintf("edit-password-%d", user.ID)),
								html.Text("New Password (leave empty to keep current)"),
							),
							html.Input(attr.Type("password").Name("password").Id(fmt.Sprintf("edit-password-%d", user.ID))),
						),
						html.Div(attr.Class("flex justify-between items-center mt-4"),
							components.Dropdown{
								Id:          "delete-user",
								ButtonText:  "Delete User",
								ButtonClass: "btn-destructive",
								Items: html.Blocks{
									html.Div(attr.Role("menuitem"),
										html.Text("Cancel"),
									),
									html.Button(attr.Form("delete-user").Type("submit").Role("menuitem").Class("text-destructive font-bold hover:bg-destructive/10"),
										html.Text("Confirm Delete"),
									),
								},
							},
							html.Div(attr.Class("flex gap-2"),
								html.Button(attr.Type("button").Class("btn-outline").Attr("onclick", "this.closest('dialog').close()"),
									html.Text("Cancel"),
								),
								html.Button(attr.Type("submit").Class("btn"),
									html.Text("Update User"),
								),
							),
						),
					),
				),
				html.Button(attr.Type("button").Attr("aria-label", "Close dialog").Attr("onclick", "this.closest('dialog').close()"),
					html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").Class("lucide lucide-x-icon lucide-x"),
						html.Elem("path", attr.Attr("d", "M18 6 6 18")),
						html.Elem("path", attr.Attr("d", "m6 6 12 12")),
					),
				),
			),
		),
		html.Script(nil, html.JS(fmt.Sprintf("document.getElementById('edit-user-dialog-%d').showModal();", user.ID))),
	), nil
}
