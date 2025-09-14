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
	Title string
	Body  html.Block
}

func (p *Page) RenderHTML() html.Block {
	return html.Blocks{
		html.Doctype("html"),
		html.Html(attr.Lang("en"),
			html.Head(nil,
				html.Meta(attr.Charset("utf-8")),
				html.Meta(attr.Name("viewport").Content("width=device-width, initial-scale=1")),
				html.Title(nil, html.Text(p.Title)),
				html.Link(attr.Href(addRefreshQuery("/dist/main.css")).Rel("stylesheet")),
				html.Script(attr.Src(addRefreshQuery("/dist/main.js")).Defer(nil)),
			),
			html.Body(nil,
				p.Body,
			),
		),
	}
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
