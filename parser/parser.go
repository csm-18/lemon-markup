package parser

import (
	"fmt"
	"strings"
	"unicode"

	"lemon/lexer"
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

type Parser struct {
	tokens    []lexer.Token
	pos       int
	templates map[string]*TemplateNode
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:    tokens,
		pos:       0,
		templates: make(map[string]*TemplateNode),
	}
}

func (p *Parser) Parse() (*Document, error) {
	doc := &Document{
		Nodes:     make([]*Node, 0),
		Templates: make(map[string]*TemplateNode),
		HasDoctype: false,
	}
	
	for !p.isAtEnd() {
		if p.peek().Type == lexer.TokenEOF {
			break
		}
		
		// Check for DOCTYPE
		if p.peek().Type == lexer.TokenText {
			text := p.peek().Value
			if strings.Contains(strings.ToUpper(text), "DOCTYPE") {
				doc.HasDoctype = true
			}
		}
		
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		
		if node != nil {
			if node.Type == NodeTemplate {
				doc.Templates[node.TemplateDef.Name] = node.TemplateDef
				p.templates[node.TemplateDef.Name] = node.TemplateDef
			} else {
				doc.Nodes = append(doc.Nodes, node)
			}
		}
	}
	
	return doc, nil
}

func (p *Parser) parseNode() (*Node, error) {
	tok := p.peek()
	
	switch tok.Type {
	case lexer.TokenText:
		p.advance()
		return &Node{
			Type:    NodeText,
			Content: tok.Value,
			Line:    tok.Line,
			Col:     tok.Col,
		}, nil
	
	case lexer.TokenVariable:
		p.advance()
		return &Node{
			Type:    NodeVariable,
			VarName: tok.Value,
			Line:    tok.Line,
			Col:     tok.Col,
		}, nil
	
	case lexer.TokenRawOpen:
		return p.parseRawBlock()
	
	case lexer.TokenTagOpen:
		return p.parseElement()
	
	case lexer.TokenEOF:
		return nil, nil
	
	default:
		return nil, fmt.Errorf("unexpected token: %v at %d:%d", tok.Type, tok.Line, tok.Col)
	}
}

func (p *Parser) parseRawBlock() (*Node, error) {
	p.advance() // consume <raw>
	
	node := &Node{
		Type:     NodeRaw,
		Children: make([]*Node, 0),
	}
	
	for !p.isAtEnd() && p.peek().Type != lexer.TokenRawClose {
		if p.peek().Type == lexer.TokenEOF {
			return nil, fmt.Errorf("unclosed raw block")
		}
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	
	if !p.isAtEnd() && p.peek().Type == lexer.TokenRawClose {
		p.advance()
	}
	
	return node, nil
}

func (p *Parser) parseElement() (*Node, error) {
	p.advance() // consume <
	
	if p.isAtEnd() || p.peek().Type != lexer.TokenTagName {
		return nil, fmt.Errorf("expected tag name")
	}
	
	tagNameTok := p.advance()
	tagName := tagNameTok.Value
	
	node := &Node{
		Type:     NodeElement,
		Tag:      tagName,
		Attrs:    make(map[string]string),
		Children: make([]*Node, 0),
		Line:     tagNameTok.Line,
		Col:      tagNameTok.Col,
	}
	
	// Check if it's a template declaration
	if tagName == "template" {
		return p.parseTemplate()
	}
	
	// Parse attributes
	for !p.isAtEnd() && p.peek().Type == lexer.TokenAttrName {
		attrName := p.advance().Value
		
		if !p.isAtEnd() && p.peek().Type == lexer.TokenAttrValue {
			attrValue := p.advance().Value
			node.Attrs[attrName] = attrValue
		} else {
			node.Attrs[attrName] = ""
		}
	}
	
	// Check for self-closing (error in Lemon Markup for custom components)
	if !p.isAtEnd() && p.peek().Type == lexer.TokenTagSelfClose {
		return nil, fmt.Errorf("self-closing tags not allowed: %s at %d:%d", tagName, node.Line, node.Col)
	}
	
	// For void elements, don't parse children
	if voidElements[tagName] {
		return node, nil
	}
	
	// Parse children until closing tag
	for !p.isAtEnd() && p.peek().Type != lexer.TokenTagClose {
		if p.peek().Type == lexer.TokenEOF {
			return nil, fmt.Errorf("unclosed tag: %s at %d:%d", tagName, node.Line, node.Col)
		}
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	
	// Consume closing tag
	if !p.isAtEnd() && p.peek().Type == lexer.TokenTagClose {
		p.advance()
		if !p.isAtEnd() && p.peek().Type == lexer.TokenTagName {
			closingTag := p.advance().Value
			if closingTag != tagName {
				return nil, fmt.Errorf("mismatched closing tag: expected %s, got %s at %d:%d", tagName, closingTag, p.peek().Line, p.peek().Col)
			}
		}
	}
	
	return node, nil
}

func (p *Parser) parseTemplate() (*Node, error) {
	// node already created with tag="template"
	// Parse attributes to get name
	attrs := make(map[string]string)
	
	for !p.isAtEnd() && p.peek().Type == lexer.TokenAttrName {
		attrName := p.advance().Value
		if !p.isAtEnd() && p.peek().Type == lexer.TokenAttrValue {
			attrValue := p.advance().Value
			attrs[attrName] = attrValue
		}
	}
	
	templateName, ok := attrs["name"]
	if !ok || templateName == "" {
		return nil, fmt.Errorf("template must have a name attribute")
	}
	
	// Validate template name
	if !unicode.IsUpper(rune(templateName[0])) {
		return nil, fmt.Errorf("template name must start with uppercase: %s", templateName)
	}
	
	reserved := map[string]bool{"template": true, "children": true, "raw": true}
	if reserved[templateName] {
		return nil, fmt.Errorf("template name '%s' is reserved", templateName)
	}
	
	// Parse template content
	content := &Node{
		Type:     NodeElement,
		Tag:      "template-content",
		Children: make([]*Node, 0),
	}
	
	for !p.isAtEnd() && p.peek().Type != lexer.TokenTagClose {
		if p.peek().Type == lexer.TokenEOF {
			return nil, fmt.Errorf("unclosed template: %s", templateName)
		}
		child, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if child != nil {
			content.Children = append(content.Children, child)
		}
	}
	
	if !p.isAtEnd() && p.peek().Type == lexer.TokenTagClose {
		p.advance()
		if !p.isAtEnd() && p.peek().Type == lexer.TokenTagName {
			closingTag := p.advance().Value
			if closingTag != "template" {
				return nil, fmt.Errorf("mismatched closing tag: expected template, got %s", closingTag)
			}
		}
	}
	
	// Extract used variables from template content
	usedVars := extractVariables(content)
	
	templateDef := &TemplateNode{
		Name:     templateName,
		Content:  content,
		UsedVars: usedVars,
	}
	
	return &Node{
		Type:        NodeTemplate,
		TemplateDef: templateDef,
	}, nil
}

func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	if !p.isAtEnd() {
		p.pos++
	}
	return tok
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens)
}

func extractVariables(node *Node) map[string]bool {
	vars := make(map[string]bool)
	if node == nil {
		return vars
	}
	
	if node.Type == NodeVariable {
		vars[node.VarName] = true
	}
	
	for _, child := range node.Children {
		childVars := extractVariables(child)
		for v := range childVars {
			vars[v] = true
		}
	}
	
	return vars
}
