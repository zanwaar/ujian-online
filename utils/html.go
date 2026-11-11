package utils

import "github.com/microcosm-cc/bluemonday"

var examHTMLPolicy = bluemonday.NewPolicy().AllowElements(
	"p",
	"br",
	"strong",
	"b",
	"em",
	"i",
	"u",
	"s",
	"ul",
	"ol",
	"li",
	"blockquote",
	"code",
	"pre",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"sub",
	"sup",
)

func SanitizeHTML(value string) string {
	return examHTMLPolicy.Sanitize(value)
}
