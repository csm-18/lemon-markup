package parser

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
	Type       NodeType
	Tag        string
	Attrs      map[string]string
	Children   []*Node
	Content    string // for text nodes
	VarName    string // for variable nodes
	TemplateDef *TemplateNode
	Line       int
	Col        int
}

type TemplateNode struct {
	Name     string
	Content  *Node
	UsedVars map[string]bool
}

type Document struct {
	Nodes        []*Node
	Templates    map[string]*TemplateNode
	HasDoctype   bool
}
