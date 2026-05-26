package generator

import (
	"strings"

	"lemon-markup/parser"
)

var voidElements = map[string]bool{
	"meta":   true,
	"link":   true,
	"img":    true,
	"br":     true,
	"hr":     true,
	"input":  true,
	"col":    true,
	"area":   true,
	"base":   true,
	"embed":  true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(doc *parser.Document) string {
	var output strings.Builder
	
	for _, node := range doc.Nodes {
		g.generateNode(node, &output)
	}
	
	return output.String()
}

func (g *Generator) generateNode(node *parser.Node, output *strings.Builder) {
	if node == nil {
		return
	}
	
	switch node.Type {
	case parser.NodeText:
		output.WriteString(node.Content)
	
	case parser.NodeVariable:
		// Variables should have been expanded, but if not, preserve them
		output.WriteString("{{ ")
		output.WriteString(node.VarName)
		output.WriteString(" }}")
	
	case parser.NodeElement:
		if node.Tag == "template-result" || node.Tag == "fragment" {
			// Don't output wrapper tags
			for _, child := range node.Children {
				g.generateNode(child, output)
			}
		} else {
			g.generateElement(node, output)
		}
	
	case parser.NodeRaw:
		for _, child := range node.Children {
			g.generateNode(child, output)
		}
	
	default:
		for _, child := range node.Children {
			g.generateNode(child, output)
		}
	}
}

func (g *Generator) generateElement(node *parser.Node, output *strings.Builder) {
	// Opening tag
	output.WriteString("<")
	output.WriteString(node.Tag)
	
	// Attributes
	for key, value := range node.Attrs {
		output.WriteString(" ")
		output.WriteString(key)
		if value != "" {
			output.WriteString("=\"")
			output.WriteString(value)
			output.WriteString("\"")
		}
	}
	
	// For void elements, don't output closing tag or children
	if voidElements[node.Tag] {
		output.WriteString(">")
		return
	}
	
	output.WriteString(">")
	
	// Children
	for _, child := range node.Children {
		g.generateNode(child, output)
	}
	
	// Closing tag
	output.WriteString("</")
	output.WriteString(node.Tag)
	output.WriteString(">")
}
