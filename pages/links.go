package pages

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
	"github.com/pkg/errors"
)

func (h *Handler) LinksPage(req *foundation.Request) (*Page, error) {
	linksFrame, err := h.LinksFrame(req)
	if err != nil {
		return nil, errors.Wrap(err, "linksFrame")
	}

	var body html.Blocks
	body.Add(linksFrame)
	body.Add(html.Elem("turbo-stream-source", attr.Src("/admin/stream/links")))

	page := &Page{
		Title:   "Quick Links - Links",
		Sidebar: Sidebar{},
		Header: Header{
			Title: "Links",
		},
		Body: body,
	}

	return page, nil
}

func (h *Handler) LinksFrame(req *foundation.Request) (html.Block, error) {
	switch req.Request.Method {
	case http.MethodPost:
		err := h.postNewLink(req)
		if err != nil {
			return nil, errors.Wrap(err, "postNewLink")
		}
	case http.MethodPatch:
		err := h.patchLink(req)
		if err != nil {
			return nil, errors.Wrap(err, "patchLink")
		}
		req.Broadcast.Send("links")
	case http.MethodDelete:
		err := h.deleteLink(req)
		if err != nil {
			return nil, errors.Wrap(err, "deleteLink")
		}
	}

	allLinks, err := h.DB.Links.AllWithVisitCounts(req.Context.Context)
	if err != nil {
		return nil, errors.Wrap(err, "AllWithVisitCounts")
	}

	return html.Elem("turbo-frame", attr.Id("links-frame"),
		html.Div(attr.Class("p-4 md:p-6 xl:p-12"),
			html.Main(attr.Class("mx-auto relative w-full max-w-screen-lg gap-10"),
				html.Div(attr.Class("flex justify-between items-center mb-4"),
					html.H2(attr.Class("text-2xl font-bold"), html.Text("Short Links")),
					newLinkForm(),
				),
				linksTable(allLinks),
			),
		),
		html.Elem("turbo-frame", attr.Id("link-dialog-frame")),
	), nil
}

func (h *Handler) postNewLink(req *foundation.Request) error {
	r := req.Request
	err := r.ParseForm()
	if err != nil {
		http.Error(req.Writer, "Failed to parse form", http.StatusBadRequest)
		return errors.Wrap(err, "ParseForm")
	}

	shortLink := r.FormValue("short_link")
	fullURL := r.FormValue("full_url")

	_, err = h.DB.Links.ByShortLink(req.Context.Context, shortLink)
	if err == nil {
		http.Error(req.Writer, fmt.Sprintf("Short link %q already exists", shortLink), http.StatusConflict)
		return errors.New("short link exists")
	}

	if req.User == nil {
		http.Error(req.Writer, "User must be logged in", http.StatusUnauthorized)
		return errors.New("not logged in")
	}

	link := &foundation.Link{
		ShortLink: shortLink,
		FullURL:   fullURL,
		UserID:    req.User.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = h.DB.Links.Insert(req.Context, link)
	if err != nil {
		return errors.Wrap(err, "Insert link")
	}

	err = req.Broadcast.Send("links")
	if err != nil {
		return errors.Wrap(err, "Broadcast.Send")
	}

	return nil
}

func (h *Handler) patchLink(req *foundation.Request) error {
	r := req.Request
	err := r.ParseForm()
	if err != nil {
		http.Error(req.Writer, "Failed to parse form", http.StatusBadRequest)
		return errors.Wrap(err, "ParseForm")
	}

	oldShortLink := req.Params.ByName("short_link")
	if oldShortLink == "" {
		http.Error(req.Writer, "Short link is required", http.StatusBadRequest)
		return errors.New("missing short link")
	}

	existingLink, err := h.DB.Links.ByShortLink(req.Context.Context, oldShortLink)
	if err != nil {
		http.Error(req.Writer, "Link not found", http.StatusNotFound)
		return errors.Wrap(err, "link not found")
	}

	newShortLink := r.FormValue("short_link")
	newFullURL := r.FormValue("full_url")

	if newShortLink == oldShortLink {
		// Just update the URL
		existingLink.FullURL = newFullURL
		existingLink.UpdatedAt = time.Now()
		err = h.DB.Links.Update(req.Context, existingLink)
		if err != nil {
			return errors.Wrap(err, "Update link")
		}
		err = req.Broadcast.Send("links")
		if err != nil {
			return errors.Wrap(err, "Broadcast.Send")
		}
	} else {
		// Check if new short link already exists
		_, err := h.DB.Links.ByShortLink(req.Context.Context, newShortLink)
		if err == nil {
			http.Error(req.Writer, fmt.Sprintf("Short link %q already exists", newShortLink), http.StatusConflict)
			return errors.New("short link exists")
		}

		// Replace the link: delete old, insert new
		err = h.DB.Links.Delete(req.Context.Context, oldShortLink)
		if err != nil {
			return errors.Wrap(err, "Delete old link")
		}

		newLink := &foundation.Link{
			ShortLink: newShortLink,
			FullURL:   newFullURL,
			UserID:    existingLink.UserID,
			CreatedAt: existingLink.CreatedAt,
			UpdatedAt: time.Now(),
		}
		err = h.DB.Links.Insert(req.Context, newLink)
		if err != nil {
			return errors.Wrap(err, "Insert new link")
		}

		err = req.Broadcast.Send("links")
		if err != nil {
			return errors.Wrap(err, "Broadcast.Send")
		}
	}

	return nil
}

func (h *Handler) deleteLink(req *foundation.Request) error {
	shortLink := req.Params.ByName("short_link")
	if shortLink == "" {
		http.Error(req.Writer, "Short link is required", http.StatusBadRequest)
		return errors.New("missing short link")
	}

	_, err := h.DB.Links.ByShortLink(req.Context.Context, shortLink)
	if err != nil {
		http.Error(req.Writer, "Link not found", http.StatusNotFound)
		return errors.Wrap(err, "link not found")
	}

	err = h.DB.Links.Delete(req.Context.Context, shortLink)
	if err != nil {
		return errors.Wrap(err, "Delete link")
	}

	err = req.Broadcast.Send("links")
	if err != nil {
		return errors.Wrap(err, "Broadcast.Send")
	}
	return nil
}

func linksTable(links []*foundation.Link) html.Block {
	var rows html.Blocks
	for _, l := range links {
		rows.Add(linkTableRow(l))
	}

	return html.Div(attr.Class("overflow-x-auto w-full"),
		html.Table(attr.Class("table"),
			html.Thead(nil,
				html.Tr(nil,
					html.Th(nil,
						html.Text("Short Link"),
					),
					html.Th(nil,
						html.Text("Full URL"),
					),
					html.Th(nil,
						html.Text("User"),
					),
					html.Th(nil,
						html.Text("Visits"),
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

func linkTableRow(link *foundation.Link) html.Block {
	displayName := "Unknown"
	if link.User != nil {
		displayName = link.User.DisplayName
	}
	return html.Tr(nil,
		html.Td(attr.Class("font-medium"),
			html.A(attr.Href(fmt.Sprintf("/%s", link.ShortLink)), html.Text(link.ShortLink)),
		),
		html.Td(nil,
			html.A(attr.Href(link.FullURL), html.Text(link.FullURL)),
		),
		html.Td(nil,
			html.Text(displayName),
		),
		html.Td(attr.Class("text-right"),
			html.Text(fmt.Sprint(link.VisitsCount)),
		),
		html.Td(attr.Class("text-right"),
			html.Text(link.CreatedAt.Format("2006-01-02 15:04")),
		),
		html.Td(attr.Class("text-right"),
			html.Text(link.UpdatedAt.Format("2006-01-02 15:04")),
		),
		html.Td(nil,
			html.A(attr.Href(fmt.Sprintf("/admin/frame/links/update/%s", link.ShortLink)).Class("btn-ghost").Attr("data-turbo-frame", "link-dialog-frame"),
				html.Text("Edit"),
			),
		),
	)
}

func newLinkForm() html.Block {
	return html.Blocks{
		html.A(attr.Href("/admin/frame/links/new").Class("btn-outline").Attr("data-turbo-frame", "link-dialog-frame"),
			html.Text("Add Link"),
		),
	}
}

func (h *Handler) LinkNewFrame(req *foundation.Request) (html.Block, error) {
	return html.Elem("turbo-frame", attr.Id("link-dialog-frame"),
		html.Dialog(attr.Id("new-link-dialog").Class("dialog w-full sm:max-w-[425px] max-h-[612px]").Attr("aria-labelledby", "new-link-dialog-title").Attr("aria-describedby", "new-link-dialog-description").Attr("onclick", "if (event.target === this) this.close()"),
			html.Article(nil,
				html.Header(nil,
					html.H2(attr.Id("new-link-dialog-title"),
						html.Text("Add New Link"),
					),
					html.P(attr.Id("new-link-dialog-description"),
						html.Text("Enter the details for the new short link."),
					),
				),
				html.Section(nil,
					html.Form(attr.Method("POST").Action("/admin/links").Class("form grid gap-4").Attr("data-turbo-frame", "links-frame"),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For("short-link"),
								html.Text("Short Link"),
							),
							html.Input(attr.Type("text").Name("short_link").Id("short-link").Required("").Autofocus("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For("full-url"),
								html.Text("Full URL"),
							),
							html.Input(attr.Type("url").Name("full_url").Id("full-url").Required("")),
						),
						html.Div(attr.Class("flex justify-end gap-2 mt-4"),
							html.Button(attr.Type("button").Class("btn-outline").Attr("onclick", "this.closest('dialog').close()"),
								html.Text("Cancel"),
							),
							html.Button(attr.Type("submit").Class("btn"),
								html.Text("Create Link"),
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
		html.Script(nil, html.JS("document.getElementById('new-link-dialog').showModal();")),
	), nil
}

func (h *Handler) LinkUpdateFrame(req *foundation.Request) (html.Block, error) {
	shortLink := req.Params.ByName("short_link")
	if shortLink == "" {
		return nil, errors.New("short link is required")
	}

	link, err := h.DB.Links.ByShortLink(req.Context.Context, shortLink)
	if err != nil {
		return nil, errors.Wrap(err, "link not found")
	}

	return html.Elem("turbo-frame", attr.Id("link-dialog-frame"),
		html.Dialog(attr.Id(fmt.Sprintf("edit-link-dialog-%s", link.ShortLink)).Class("dialog w-full sm:max-w-[425px] max-h-[612px]").Attr("aria-labelledby", fmt.Sprintf("edit-link-dialog-title-%s", link.ShortLink)).Attr("aria-describedby", fmt.Sprintf("edit-link-dialog-description-%s", link.ShortLink)).Attr("onclick", "if (event.target === this) this.close()"),
			html.Article(nil,
				html.Header(nil,
					html.H2(attr.Id(fmt.Sprintf("edit-link-dialog-title-%s", link.ShortLink)),
						html.Text("Edit Link"),
					),
					html.P(attr.Id(fmt.Sprintf("edit-link-dialog-description-%s", link.ShortLink)),
						html.Text("Make changes to the link details."),
					),
				),
				html.Section(nil,
					html.Form(attr.Method("PATCH").Action(fmt.Sprintf("/admin/links/%s", link.ShortLink)).Class("form grid gap-4").Attr("data-turbo-frame", "links-frame"),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For(fmt.Sprintf("edit-short-link-%s", link.ShortLink)),
								html.Text("Short Link"),
							),
							html.Input(attr.Type("text").Name("short_link").Id(fmt.Sprintf("edit-short-link-%s", link.ShortLink)).Value(link.ShortLink).Required("")),
						),
						html.Div(attr.Class("grid gap-3"),
							html.Label(attr.For(fmt.Sprintf("edit-full-url-%s", link.ShortLink)),
								html.Text("Full URL"),
							),
							html.Input(attr.Type("url").Name("full_url").Id(fmt.Sprintf("edit-full-url-%s", link.ShortLink)).Value(link.FullURL).Required("").Autofocus("")),
						),
						html.Div(attr.Class("flex justify-end gap-2 mt-4"),
							html.Button(attr.Type("button").Class("btn-outline").Attr("onclick", "this.closest('dialog').close()"),
								html.Text("Cancel"),
							),
							html.Button(attr.Type("submit").Class("btn"),
								html.Text("Update Link"),
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
		html.Script(nil, html.JS(fmt.Sprintf("document.getElementById('edit-link-dialog-%s').showModal();", link.ShortLink))),
	), nil
}

func (h *Handler) ShortLinkHandler(req *foundation.Request) (html.Block, error) {
	path := strings.TrimPrefix(req.Request.URL.Path, "/")

	link, err := h.DB.Links.ByShortLink(req.Context.Context, path)
	if err != nil || link == nil {
		req.Writer.WriteHeader(http.StatusNotFound)
		return html.Text("page not found"), errors.New("not found")
	}

	visit := &foundation.LinkVisit{
		ShortLink: path,
		UserID:    req.Session.UserID,
	}
	err = h.DB.Visits.Insert(req.Context.Context, visit)
	if err != nil {
		return nil, errors.Wrap(err, "Visits.Insert")
	}

	err = req.Broadcast.Send("links")
	if err != nil {
		return nil, errors.Wrap(err, "Broadcast.Send")
	}

	http.Redirect(req.Writer, req.Request, link.FullURL, http.StatusFound)
	return nil, nil
}

func (h *Handler) LinksStream(req *foundation.Request) (html.Block, error) {
	frame, err := h.LinksFrame(req)
	if err != nil {
		return nil, errors.Wrap(err, "LinksFrame")
	}

	return html.Elem("turbo-stream", attr.Action("replace").Target("links-frame"),
		html.Template(nil, frame)), nil
}
