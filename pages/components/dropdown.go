package components

import (
	"fmt"

	"github.com/mbertschler/html"
	"github.com/mbertschler/html/attr"
)

type Dropdown struct {
	Id          string
	ButtonClass string
	ButtonText  string
	Items       html.Blocks
}

func (d Dropdown) RenderHTML() html.Block {
	return html.Div(attr.Id(d.Id).Class("dropdown-menu"),
		html.Button(attr.Type("button").Id(fmt.Sprintf("%s-trigger", d.Id)).Attr("aria-haspopup", "menu").Attr("aria-controls", fmt.Sprintf("%s-menu", d.Id)).Attr("aria-expanded", "false").Class(d.ButtonClass),
			html.Text(d.ButtonText),
		),
		html.Div(attr.Id(fmt.Sprintf("%s-popover", d.Id)).Attr("data-popover", nil).Attr("aria-hidden", "true").Class("min-w-56"),
			html.Div(attr.Role("menu").Id(fmt.Sprintf("%s-menu", d.Id)).Attr("aria-labelledby", fmt.Sprintf("%s-trigger", d.Id)),
				d.Items,
			),
		),
	)
}
