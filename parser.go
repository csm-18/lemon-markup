package main

import (
	"fmt"
	"os"
	"unicode"
)

// Standard HTML Void Elements list used to prevent parsing children for self-closing native tags
var voidElements = map[string]bool{
	"meta": true, "link": true, "img": true, "br": true, "hr": true,
	"input": true, "col": true, "area": true, "base": true, "embed": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

type NodeType int

const (
	NodeDocument NodeType = iota
	NodeElement
	NodeText
	NodeVariable
	NodeTemplate
	NodeRaw
)

type Node struct {
	Type        NodeType
	Tag         string
	Attrs       map[string]string
	Children    []*Node
	Content     string // For plain text or doctype content
	VarName     string // For {{ Variable }} keys
	TemplateDef *TemplateNode
	Line        int
	Col         int
}

type TemplateNode struct {
	Name     string
	Content  *Node
	UsedVars map[string]bool
}

type Document struct {
	Nodes      []*Node
	Templates  map[string]*TemplateNode
	HasDoctype bool
}

type Parser struct {
	tokens   []Token
	pos      int
	fileName string
}

func NewParser(tokens []Token, fileName string) *Parser {
	return &Parser{
		tokens:   tokens,
		pos:      0,
		fileName: fileName,
	}
}

// Global Panic-Exit Helper to match the strict diagnostics rule of the Lexer
func (p *Parser) fatalError(msg string, line, col int) {
	fmt.Printf("Syntax Error: %s\n", msg)
	fmt.Printf("  --> %s:%d:%d\n", p.fileName, line, col)
	os.Exit(0)
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) Parse() *Document {
	doc := &Document{
		Nodes:      make([]*Node, 0),
		Templates:  make(map[string]*TemplateNode),
		HasDoctype: false,
	}

	for p.peek().Type != TokenEOF {
		tok := p.peek()

		// Direct Doctype Detection leveraging your manual Lexer token
		if tok.Type == TokenDocTypeDecl {
			doc.HasDoctype = true
			p.advance()
			doc.Nodes = append(doc.Nodes, &Node{
				Type:    NodeText,
				Content: tok.Value,
				Line:    tok.Line,
				Col:     tok.Col,
			})
			continue
		}

		node := p.parseNode()
		if node != nil {
			if node.Type == NodeTemplate {
				// Global Registry Collision Guard
				if _, exists := doc.Templates[node.TemplateDef.Name]; exists {
					p.fatalError(fmt.Sprintf("Duplicate template name '%s' declared", node.TemplateDef.Name), node.Line, node.Col)
				}
				doc.Templates[node.TemplateDef.Name] = node.TemplateDef
			} else {
				doc.Nodes = append(doc.Nodes, node)
			}
		}
	}

	return doc
}

func (p *Parser) parseNode() *Node {
	tok := p.peek()

	switch tok.Type {
	case TokenText:
		p.advance()
		return &Node{
			Type:    NodeText,
			Content: tok.Value,
			Line:    tok.Line,
			Col:     tok.Col,
		}

	case TokenVariable:
		p.advance()
		return &Node{
			Type:    NodeVariable,
			VarName: tok.Value,
			Line:    tok.Line,
			Col:     tok.Col,
		}

	case TokenRawOpen:
		return p.parseRawBlock()

	case TokenTagOpen:
		return p.parseElement()

	default:
		p.fatalError(fmt.Sprintf("Unexpected token outside structural block elements"), tok.Line, tok.Col)
		return nil
	}
}

func (p *Parser) parseRawBlock() *Node {
	startTok := p.advance() // Consume TokenRawOpen (<raw>)

	node := &Node{
		Type:     NodeRaw,
		Children: make([]*Node, 0),
		Line:     startTok.Line,
		Col:      startTok.Col,
	}

	for p.peek().Type != TokenRawClose {
		if p.peek().Type == TokenEOF {
			p.fatalError("Unclosed literal escape block. Missing matching '</raw>'", startTok.Line, startTok.Col)
		}
		child := p.parseNode()
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	p.advance() // Consume TokenRawClose (</raw>)
	return node
}

func (p *Parser) parseElement() *Node {
	openTok := p.advance() // Consume TokenTagOpen (<)

	if p.peek().Type != TokenTagName {
		p.fatalError("Expected a valid structural tag identifier name directly following opening bracket '<'", openTok.Line, openTok.Col)
	}

	nameTok := p.advance()
	tagName := nameTok.Value

	// Intercept Template Blocks immediately
	if tagName == "template" {
		return p.parseTemplate(nameTok.Line, nameTok.Col)
	}

	node := &Node{
		Type:     NodeElement,
		Tag:      tagName,
		Attrs:    make(map[string]string),
		Children: make([]*Node, 0),
		Line:     nameTok.Line,
		Col:      nameTok.Col,
	}

	// Read Attributes / Props stream
	for p.peek().Type == TokenAttrName {
		attrName := p.advance().Value
		if p.peek().Type == TokenAttrValue {
			node.Attrs[attrName] = p.advance().Value
		} else {
			node.Attrs[attrName] = ""
		}
	}

	// Void elements exit early cleanly without looking for an explicit structural close block
	if voidElements[tagName] {
		return node
	}

	// Collect inner structural children nodes until matching close sequence tag is found
	for p.peek().Type != TokenTagClose {
		if p.peek().Type == TokenEOF {
			p.fatalError(fmt.Sprintf("Unclosed markup block container tag element. Missing matching closing '</%s>'", tagName), node.Line, node.Col)
		}
		child := p.parseNode()
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}

	p.advance() // Consume TokenTagClose (</)

	if p.peek().Type != TokenTagName {
		p.fatalError(fmt.Sprintf("Expected explicit matching closing tag name '%s' following closing branch identifier", tagName), p.peek().Line, p.peek().Col)
	}

	closingNameTok := p.advance()
	if closingNameTok.Value != tagName {
		p.fatalError(fmt.Sprintf("Mismatched block closing tag structure. Expected '</%s>', but encountered '</%s>' alternative instead", tagName, closingNameTok.Value), closingNameTok.Line, closingNameTok.Col)
	}

	return node
}

func (p *Parser) parseTemplate(startLine, startCol int) *Node {
	attrs := make(map[string]string)

	for p.peek().Type == TokenAttrName {
		attrName := p.advance().Value
		if p.peek().Type == TokenAttrValue {
			attrs[attrName] = p.advance().Value
		}
	}

	templateName, ok := attrs["name"]
	if !ok || templateName == "" {
		p.fatalError("Component template declaration definitions are strictly required to define a valid identifier 'name' string key", startLine, startCol)
	}

	// Spec rule verification check: PascalCase verification
	if !unicode.IsUpper(rune(templateName[0])) {
		p.fatalError(fmt.Sprintf("Custom Component template identifiers must start with an uppercase letter character choice: '%s'", templateName), startLine, startCol)
	}

	// Spec rule verification check: Reserved values check
	reserved := map[string]bool{"template": true, "children": true, "raw": true}
	if reserved[templateName] {
		p.fatalError(fmt.Sprintf("Template naming identifier token selection '%s' is an immutable reserved keyword component element value", templateName), startLine, startCol)
	}

	contentRoot := &Node{
		Type:     NodeElement,
		Tag:      "template-content",
		Children: make([]*Node, 0),
		Line:     startLine,
		Col:      startCol,
	}

	// Read inner core template nodes logic loop
	for p.peek().Type != TokenTagClose {
		if p.peek().Type == TokenEOF {
			p.fatalError(fmt.Sprintf("Unterminated template structure definition block layer. Missing closing structural '</template>' for template identifier: '%s'", templateName), startLine, startCol)
		}

		// Spec constraint verification check: Nesting verification
		if p.peek().Type == TokenTagOpen {
			// Look ahead behind the bracket marker
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Value == "template" {
				p.fatalError("Component template block architectural nesting declarations are strictly forbidden inside alternative component layout scopes", p.peek().Line, p.peek().Col)
			}
		}

		child := p.parseNode()
		if child != nil {
			contentRoot.Children = append(contentRoot.Children, child)
		}
	}

	p.advance() // Consume TokenTagClose (</)

	closingNameTok := p.advance()
	if closingNameTok.Value != "template" {
		p.fatalError(fmt.Sprintf("Mismatched component scope wrapper structure boundary. Expected outer layout close code sequence '</template>', but received '</%s>' alternative", closingNameTok.Value), closingNameTok.Line, closingNameTok.Col)
	}

	usedVars := make(map[string]bool)
	extractVariables(contentRoot, usedVars)

	return &Node{
		Type: NodeTemplate,
		Line: startLine,
		Col:  startCol,
		TemplateDef: &TemplateNode{
			Name:     templateName,
			Content:  contentRoot,
			UsedVars: usedVars,
		},
	}
}

// Recursive Variable Extractor Function
func extractVariables(node *Node, vars map[string]bool) {
	if node == nil {
		return
	}
	if node.Type == NodeVariable {
		vars[node.VarName] = true
	}
	for _, child := range node.Children {
		extractVariables(child, vars)
	}
}
