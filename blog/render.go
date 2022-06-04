package blog

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

func RenderToHtmlBytes(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.FencedCode
	parser := parser.NewWithExtensions(extensions)
	output := markdown.ToHTML(md, parser, nil)
	return output
}

func RenderToHtmlStr(md []byte) string {
	output := RenderToHtmlBytes(md)
	return string(output)
}

func RenderToHtml(md []byte) template.HTML {
	output := RenderToHtmlBytes(md)
	return template.HTML(output)
}
