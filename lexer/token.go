package lexer

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenText
	TokenTagOpen
	TokenTagClose
	TokenTagSelfClose
	TokenTagName
	TokenAttrName
	TokenAttrValue
	TokenVariable
	TokenRawOpen
	TokenRawClose
	TokenError
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}
