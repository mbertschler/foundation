package pages

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

type PageFunc func(*foundation.Request) (*Page, error)
type FrameFunc func(*foundation.Request) (html.Block, error)

type Page struct {
	Title   string
	Sidebar html.Block
	Header  html.Block
	Body    html.Block
}

func (p *Page) RenderHTML(req *foundation.Request) html.Block {
	return html.Blocks{
		html.Doctype("html"),
		html.Html(attr.Lang("en"),
			html.Head(nil,
				html.Meta(attr.Charset("utf-8")),
				html.Meta(attr.Name("viewport").Content("width=device-width, initial-scale=1")),
				html.Meta(attr.Name("csrf-token").Content(req.CSRFToken())),
				html.Title(nil, html.Text(p.Title)),
				html.Link(attr.Href(addRefreshQuery("/dist/main.css")).Rel("stylesheet")),
				html.Script(attr.Src(addRefreshQuery("/dist/main.js")).Defer(nil)),
			),
			html.Body(nil,
				p.Sidebar,
				html.Main(nil,
					p.Header,
					p.Body,
				),
			),
		),
	}
}

type Sidebar struct {
}

func (s Sidebar) RenderHTML() html.Block {
	return html.Aside(attr.Class("sidebar").DataAttr("side", "left").Attr("aria-hidden", "false"),
		html.Nav(attr.Attr("aria-label", "Sidebar navigation"),
			html.Section(attr.Class("scrollbar"),
				html.Div(attr.Role("group").Attr("aria-labelledby", "group-label-content-1").Class("flex flex-col h-full"),
					html.H1(attr.Id("group-label-content-1").Class("text-2xl font-semibold tracking-tight m-2 mb-4"),
						html.Text("Quicklink"),
					),
					html.Ul(attr.Class("flex-grow"),
						html.Li(nil,
							html.A(attr.Href("/admin"),
								html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
									html.Elem("path", attr.Attr("d", "m7 11 2-2-2-2")),
									html.Elem("path", attr.Attr("d", "M11 13h4")),
									html.Elem("rect", attr.Width("18").Height("18").Attr("x", "3").Attr("y", "3").Attr("rx", "2").Attr("ry", "2")),
								),
								html.Span(nil,
									html.Text("Links"),
								),
							),
						),
						html.Li(nil,
							html.A(attr.Href("/admin/users"),
								html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
									html.Elem("path", attr.Attr("d", "M12 8V4H8")),
									html.Elem("rect", attr.Width("16").Height("12").Attr("x", "4").Attr("y", "8").Attr("rx", "2")),
									html.Elem("path", attr.Attr("d", "M2 14h2")),
									html.Elem("path", attr.Attr("d", "M20 14h2")),
									html.Elem("path", attr.Attr("d", "M15 13v2")),
									html.Elem("path", attr.Attr("d", "M9 13v2")),
								),
								html.Span(nil,
									html.Text("Users"),
								),
							),
						),
					),
					html.Form(attr.Class("mt-auto mb-2 text-center").Method("POST").Action("/admin/logout"),
						html.Button(attr.Class("btn-outline").Type("submit"),
							html.Text("Logout"),
						),
					),
				),
			),
		),
	)

}

type Header struct {
	Title string
}

func (h Header) RenderHTML() html.Block {
	return html.Header(attr.Class("bg-background sticky inset-x-0 top-0 isolate flex shrink-0 items-center gap-2 border-b z-10"),
		html.Div(attr.Class("flex h-14 w-full items-center gap-2 px-4"),
			html.Button(attr.Type("button").Attr("onclick", "document.dispatchEvent(new CustomEvent('basecoat:sidebar'))").Attr("aria-label", "Toggle sidebar").DataAttr("tooltip", "Toggle sidebar").DataAttr("side", "bottom").DataAttr("align", "start").Class("btn-sm-icon-ghost size-7 -ml-1.5"),
				html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
					html.Elem("rect", attr.Width("18").Height("18").Attr("x", "3").Attr("y", "3").Attr("rx", "2")),
					html.Elem("path", attr.Attr("d", "M9 3v18")),
				),
			),
			html.H2(attr.Id("group-label-content-1").Class("text-xl font-semibold tracking-tight mr-auto"),
				html.Text(h.Title),
			),
			html.Button(attr.Type("button").Attr("aria-label", "Toggle dark mode").DataAttr("tooltip", "Toggle dark mode").DataAttr("side", "bottom").Attr("onclick", "document.dispatchEvent(new CustomEvent('basecoat:theme'))").Class("btn-icon-outline size-8"),
				html.Span(attr.Class("hidden dark:block"),
					html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
						html.Elem("circle", attr.Attr("cx", "12").Attr("cy", "12").Attr("r", "4")),
						html.Elem("path", attr.Attr("d", "M12 2v2")),
						html.Elem("path", attr.Attr("d", "M12 20v2")),
						html.Elem("path", attr.Attr("d", "m4.93 4.93 1.41 1.41")),
						html.Elem("path", attr.Attr("d", "m17.66 17.66 1.41 1.41")),
						html.Elem("path", attr.Attr("d", "M2 12h2")),
						html.Elem("path", attr.Attr("d", "M20 12h2")),
						html.Elem("path", attr.Attr("d", "m6.34 17.66-1.41 1.41")),
						html.Elem("path", attr.Attr("d", "m19.07 4.93-1.41 1.41")),
					),
				),
				html.Span(attr.Class("block dark:hidden"),
					html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
						html.Elem("path", attr.Attr("d", "M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z")),
					),
				),
			),
			html.A(attr.Href("https://github.com/mbertschler/foundation").Class("btn-icon size-8").Target("_blank").Rel("noopener noreferrer").DataAttr("tooltip", "GitHub repository").DataAttr("side", "bottom").DataAttr("align", "end"),
				html.Elem("svg", attr.Attr("xmlns", "http://www.w3.org/2000/svg").Width("24").Height("24").Attr("viewbox", "0 0 24 24").Attr("fill", "none").Attr("stroke", "currentColor").Attr("stroke-width", "2").Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round"),
					html.Elem("path", attr.Attr("d", "M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4")),
					html.Elem("path", attr.Attr("d", "M9 18c-4.51 2-5-2-7-2")),
				),
			),
		),
	)
}

var startupTime = time.Now()

func addRefreshQuery(in string) string {
	return fmt.Sprint(in, "?t=", int64ToURLSafeString(startupTime.Unix()))
}

func int64ToURLSafeString(n int64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))

	// Remove leading zeros to make it shorter
	start := 0
	for start < len(buf) && buf[start] == 0 {
		start++
	}
	if start == len(buf) {
		start = len(buf) - 1 // Handle zero case
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf[start:])
}
