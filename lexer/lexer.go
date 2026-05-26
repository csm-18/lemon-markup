package lexer

import (
	"strings"
	"unicode"
)

type Lexer struct {
	input   string
	pos     int
	line    int
	col     int
	inRaw   bool
	tokens  []Token
	textBuf strings.Builder
}

func New(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		if l.inRaw {
			// When in raw mode, consume all text until </raw>
			if l.checkString(0, "</raw>") {
				// Process the closing raw tag
				l.flushText()
				l.advance() // <
				l.advance() // /
				l.advance() // r
				l.advance() // a
				l.advance() // w
				l.advance() // >
				l.inRaw = false
				l.tokens = append(l.tokens, Token{
					Type:  TokenRawClose,
					Value: "raw",
					Line:  l.line,
					Col:   l.col,
				})
			} else {
				// Read raw text
				ch := l.peek()
				l.textBuf.WriteByte(ch)
				l.advance()
			}
		} else {
			ch := l.peek()
			if ch == '<' && l.isTagStart() {
				l.flushText()
				l.readTag()
			} else if ch == '{' && l.peekAhead(1) == '{' {
				l.flushText()
				l.readVariable()
			} else {
				l.textBuf.WriteByte(ch)
				l.advance()
			}
		}
	}
	l.flushText()
	l.tokens = append(l.tokens, Token{Type: TokenEOF})
	return l.tokens
}

func (l *Lexer) isTagStart() bool {
	// Check if next chars form a tag opening
	if l.pos+1 >= len(l.input) {
		return false
	}
	next := l.input[l.pos+1]
	return next == '/' || next == '!' || unicode.IsLetter(rune(next))
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekAhead(n int) byte {
	pos := l.pos + n
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *Lexer) flushText() {
	if l.textBuf.Len() > 0 {
		l.tokens = append(l.tokens, Token{
			Type:  TokenText,
			Value: l.textBuf.String(),
			Line:  l.line,
			Col:   l.col,
		})
		l.textBuf.Reset()
	}
}

func (l *Lexer) readVariable() {
	startLine := l.line
	startCol := l.col
	l.advance() // {
	l.advance() // {
	
	var varBuf strings.Builder
	for l.pos < len(l.input) {
		if l.peek() == '}' && l.peekAhead(1) == '}' {
			l.advance()
			l.advance()
			break
		}
		varBuf.WriteByte(l.peek())
		l.advance()
	}
	
	l.tokens = append(l.tokens, Token{
		Type:  TokenVariable,
		Value: strings.TrimSpace(varBuf.String()),
		Line:  startLine,
		Col:   startCol,
	})
}

func (l *Lexer) readTag() {
	startLine := l.line
	startCol := l.col
	l.advance() // <
	
	// Handle DOCTYPE
	if l.peek() == '!' {
		var buf strings.Builder
		for l.pos < len(l.input) && l.peek() != '>' {
			buf.WriteByte(l.peek())
			l.advance()
		}
		if l.peek() == '>' {
			l.advance()
		}
		// Store DOCTYPE as text
		l.tokens = append(l.tokens, Token{
			Type:  TokenText,
			Value: "<" + buf.String() + ">",
			Line:  startLine,
			Col:   startCol,
		})
		return
	}
	
	// Check for opening raw tag <raw>
	if l.checkString(0, "raw>") {
		l.inRaw = true
		l.advance() // r
		l.advance() // a
		l.advance() // w
		l.advance() // >
		l.tokens = append(l.tokens, Token{
			Type:  TokenRawOpen,
			Value: "raw",
			Line:  startLine,
			Col:   startCol,
		})
		return
	}
	
	// Check for closing tag
	if l.peek() == '/' {
		l.advance()
		l.tokens = append(l.tokens, Token{
			Type:  TokenTagClose,
			Value: "",
			Line:  startLine,
			Col:   startCol,
		})
		tagName := l.readTagName()
		l.skipWhitespace()
		if l.peek() != '>' {
			l.tokens = append(l.tokens, Token{Type: TokenError, Value: "Expected >", Line: l.line, Col: l.col})
			return
		}
		l.advance() // >
		l.tokens = append(l.tokens, Token{
			Type:  TokenTagName,
			Value: tagName,
			Line:  startLine,
			Col:   startCol,
		})
		return
	}
	
	// Opening tag
	tagName := l.readTagName()
	l.skipWhitespace()
	
	l.tokens = append(l.tokens, Token{
		Type:  TokenTagOpen,
		Value: "",
		Line:  startLine,
		Col:   startCol,
	})
	l.tokens = append(l.tokens, Token{
		Type:  TokenTagName,
		Value: tagName,
		Line:  startLine,
		Col:   startCol,
	})
	
	// Read attributes
	for l.peek() != '>' && l.peek() != '/' && l.pos < len(l.input) {
		l.skipWhitespace()
		if l.peek() == '>' || l.peek() == '/' {
			break
		}
		
		attrName := l.readAttrName()
		if attrName == "" {
			break
		}
		
		l.tokens = append(l.tokens, Token{
			Type:  TokenAttrName,
			Value: attrName,
			Line:  l.line,
			Col:   l.col,
		})
		
		l.skipWhitespace()
		if l.peek() == '=' {
			l.advance()
			l.skipWhitespace()
			attrValue := l.readAttrValue()
			l.tokens = append(l.tokens, Token{
				Type:  TokenAttrValue,
				Value: attrValue,
				Line:  l.line,
				Col:   l.col,
			})
		}
	}
	
	// Check for self-closing (not allowed in Lemon Markup)
	if l.peek() == '/' {
		l.advance()
		if l.peek() == '>' {
			l.advance()
			l.tokens = append(l.tokens, Token{
				Type:  TokenTagSelfClose,
				Value: "",
				Line:  startLine,
				Col:   startCol,
			})
			return
		}
	}
	
	if l.peek() == '>' {
		l.advance()
	}
}

func (l *Lexer) readTagName() string {
	var name strings.Builder
	for l.peek() != 0 && (unicode.IsLetter(rune(l.peek())) || unicode.IsDigit(rune(l.peek()))) {
		name.WriteByte(l.peek())
		l.advance()
	}
	return name.String()
}

func (l *Lexer) readAttrName() string {
	var name strings.Builder
	for l.peek() != 0 && (unicode.IsLetter(rune(l.peek())) || unicode.IsDigit(rune(l.peek())) || l.peek() == '-' || l.peek() == '_') {
		name.WriteByte(l.peek())
		l.advance()
	}
	return name.String()
}

func (l *Lexer) readAttrValue() string {
	var value strings.Builder
	
	if l.peek() == '"' {
		l.advance()
		for l.peek() != 0 && l.peek() != '"' {
			if l.peek() == '\\' && l.peekAhead(1) == '"' {
				l.advance()
				value.WriteByte(l.peek())
				l.advance()
			} else {
				value.WriteByte(l.peek())
				l.advance()
			}
		}
		if l.peek() == '"' {
			l.advance()
		}
	} else if l.peek() == '\'' {
		l.advance()
		for l.peek() != 0 && l.peek() != '\'' {
			value.WriteByte(l.peek())
			l.advance()
		}
		if l.peek() == '\'' {
			l.advance()
		}
	}
	
	return value.String()
}

func (l *Lexer) checkString(offset int, s string) bool {
	for i := 0; i < len(s); i++ {
		if l.peekAhead(offset+i) != s[i] {
			return false
		}
	}
	return true
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\n' || l.peek() == '\r') {
		l.advance()
	}
}
