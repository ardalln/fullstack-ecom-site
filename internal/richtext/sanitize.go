package richtext

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var allowedElements = map[string]bool{
	"a": true, "b": true, "blockquote": true, "br": true, "div": true, "em": true,
	"h2": true, "h3": true, "h4": true, "hr": true, "i": true, "li": true,
	"ol": true, "p": true, "s": true, "strike": true, "strong": true,
	"sub": true, "sup": true, "u": true, "ul": true,
}

var discardedContent = map[string]bool{
	"iframe": true, "object": true, "script": true, "style": true,
	"svg": true, "math": true, "template": true,
}

// Sanitize keeps a small semantic formatting allowlist and removes scripts,
// event handlers, styles, and unsafe link schemes before content is stored.
func Sanitize(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	contextNode := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(input), contextNode)
	if err != nil {
		return html.EscapeString(input)
	}
	var output strings.Builder
	for _, node := range nodes {
		for _, safe := range cleanNode(node) {
			if err := html.Render(&output, safe); err != nil {
				return html.EscapeString(input)
			}
		}
	}
	return output.String()
}

// PlainText returns the text content of sanitized rich text.
func PlainText(input string) string {
	safe := Sanitize(input)
	contextNode := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(safe), contextNode)
	if err != nil {
		return strings.TrimSpace(input)
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
			text.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	for _, node := range nodes {
		visit(node)
	}
	return strings.Join(strings.Fields(text.String()), " ")
}

func cleanNode(node *html.Node) []*html.Node {
	switch node.Type {
	case html.TextNode:
		return []*html.Node{{Type: html.TextNode, Data: node.Data}}
	case html.ElementNode:
		tag := strings.ToLower(node.Data)
		if discardedContent[tag] {
			return nil
		}
		children := make([]*html.Node, 0, 2)
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			children = append(children, cleanNode(child)...)
		}
		if !allowedElements[tag] {
			return children
		}
		safe := &html.Node{Type: html.ElementNode, Data: tag, DataAtom: atom.Lookup([]byte(tag))}
		if tag == "a" {
			for _, attr := range node.Attr {
				switch strings.ToLower(attr.Key) {
				case "href":
					if safeHref(attr.Val) {
						safe.Attr = append(safe.Attr, html.Attribute{Key: "href", Val: strings.TrimSpace(attr.Val)})
					}
				case "title":
					if len(attr.Val) <= 200 {
						safe.Attr = append(safe.Attr, html.Attribute{Key: "title", Val: attr.Val})
					}
				case "target":
					if attr.Val == "_blank" || attr.Val == "_self" {
						safe.Attr = append(safe.Attr, html.Attribute{Key: "target", Val: attr.Val})
						if attr.Val == "_blank" {
							safe.Attr = append(safe.Attr, html.Attribute{Key: "rel", Val: "noopener noreferrer"})
						}
					}
				}
			}
		}
		for _, child := range children {
			safe.AppendChild(child)
		}
		return []*html.Node{safe}
	default:
		return nil
	}
}

func safeHref(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 2048 || strings.Contains(value, `\`) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return false
	}
	// Parse URL schemes without accepting HTML or script syntax as a link.
	colon := strings.IndexByte(value, ':')
	separators := strings.IndexAny(value, "/?#")
	if colon >= 0 && (separators < 0 || colon < separators) {
		scheme := strings.ToLower(value[:colon])
		return scheme == "http" || scheme == "https" || scheme == "mailto" || scheme == "tel"
	}
	return !strings.HasPrefix(value, "//")
}
