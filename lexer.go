package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenDocTypeDecl
	TokenText
	TokenTagOpen
	TokenTagClose
	TokenTagName
	TokenAttrName
	TokenAttrValue
	TokenVariable
	TokenRawOpen
	TokenRawClose
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func Lexer(srcFile File) []Token {
	var tokens []Token
	text := srcFile.text

	x := 0
	line := 1
	col := 1
	inRaw := false

	// Helper to advance the global index pointer while strictly tracking line & column counters
	advance := func(n int) {
		for i := 0; i < n; i++ {
			if x >= len(text) {
				break
			}
			if text[x] == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			x++
		}
	}

	// Buffer to cleanly collect continuous blocks of standard text/whitespace contents
	var textBuf strings.Builder
	flushText := func() {
		if textBuf.Len() > 0 {
			tokens = append(tokens, Token{
				Type:  TokenText,
				Value: textBuf.String(),
				Line:  line,
				Col:   col,
			})
			textBuf.Reset()
		}
	}

	for x < len(text) {
		if inRaw {
			// Inside raw text literal extraction zone, look strictly for closing structural segment
			if strings.HasPrefix(text[x:], "</raw>") {
				flushText()
				tokens = append(tokens, Token{Type: TokenRawClose, Value: "</raw>", Line: line, Col: col})
				advance(6)
				inRaw = false
			} else {
				textBuf.WriteByte(text[x])
				advance(1)
			}
		} else {
			// 1. DYNAMIC EXPRESSION TEMPLATE MARKERS ({{ Expression }})
			if strings.HasPrefix(text[x:], "{{") {
				flushText()
				startLine := line
				startCol := col
				advance(2)

				var varBuf strings.Builder
				foundEnd := false

				for x < len(text) {
					if strings.HasPrefix(text[x:], "}}") {
						foundEnd = true
						advance(2)
						break
					}
					varBuf.WriteByte(text[x])
					advance(1)
				}

				if !foundEnd {
					fmt.Printf("Syntax Error: Missing ending variable delimiter '}}' wrapping layout token\n")
					fmt.Printf("  --> %s:%d:%d\n", srcFile.name, startLine, startCol)
					os.Exit(0)
				}

				tokens = append(tokens, Token{
					Type:  TokenVariable,
					Value: strings.TrimSpace(varBuf.String()),
					Line:  startLine,
					Col:   startCol,
				})
				continue
			}

			// 2. MARKUP COMPONENT TAG STRUCT BOUNDARIES (<tag, </tag, <!DOCTYPE)
			if text[x] == '<' {
				flushText()
				startLine := line
				startCol := col

				// Check for DOCTYPE declaration (case-insensitive check for <!doctype or <!DOCTYPE)
				if x+9 <= len(text) && strings.ToLower(text[x:x+9]) == "<!doctype" {
					var docTypeBuf strings.Builder
					docTypeLine := line
					docTypeCol := col

					// Read until the matching closing '>' is found
					foundEnd := false
					for x < len(text) {
						docTypeBuf.WriteByte(text[x])
						if text[x] == '>' {
							advance(1)
							foundEnd = true
							break
						}
						advance(1)
					}

					if !foundEnd {
						fmt.Printf("Syntax Error: Missing closing delimiter '>' for DOCTYPE definition\n")
						fmt.Printf("  --> %s:%d:%d\n", srcFile.name, docTypeLine, docTypeCol)
						os.Exit(0)
					}

					tokens = append(tokens, Token{
						Type:  TokenDocTypeDecl,
						Value: docTypeBuf.String(),
						Line:  docTypeLine,
						Col:   docTypeCol,
					})
					continue
				}

				// Check for raw literal escaping sections blocks
				if strings.HasPrefix(text[x:], "<raw>") {
					tokens = append(tokens, Token{Type: TokenRawOpen, Value: "<raw>", Line: line, Col: col})
					advance(5)
					inRaw = true
					continue
				}

				// Check for traditional XML closing segment sequences
				if x+1 < len(text) && text[x+1] == '/' {
					tokens = append(tokens, Token{Type: TokenTagClose, Value: "</", Line: line, Col: col})
					advance(2)

					// Extract Tag name trailing marker content safely
					var nameBuf strings.Builder
					nameLine := line
					nameCol := col
					for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x])) || text[x] == '-' || text[x] == '_') {
						nameBuf.WriteByte(text[x])
						advance(1)
					}

					if nameBuf.Len() > 0 {
						tokens = append(tokens, Token{
							Type:  TokenTagName,
							Value: nameBuf.String(),
							Line:  nameLine,
							Col:   nameCol,
						})
					}

					// Consume spaces until matching bracket closure is successfully reached
					for x < len(text) && unicode.IsSpace(rune(text[x])) {
						advance(1)
					}

					if x < len(text) && text[x] == '>' {
						advance(1)
					} else {
						fmt.Printf("Syntax Error: Missing closing delimiter '>' for closing element layout definition\n")
						fmt.Printf("  --> %s:%d:%d\n", srcFile.name, startLine, startCol)
						os.Exit(0)
					}
					continue
				} else {
					// Fallthrough: Handle basic element tag open layer strings
					tokens = append(tokens, Token{Type: TokenTagOpen, Value: "<", Line: line, Col: col})
					advance(1)

					// Extract element name identifier layer
					var nameBuf strings.Builder
					nameLine := line
					nameCol := col
					for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x])) || text[x] == '-' || text[x] == '_') {
						nameBuf.WriteByte(text[x])
						advance(1)
					}

					if nameBuf.Len() > 0 {
						tokens = append(tokens, Token{
							Type:  TokenTagName,
							Value: nameBuf.String(),
							Line:  nameLine,
							Col:   nameCol,
						})
					}

					tagClosed := false
					for x < len(text) {
						// Strip interior spacer layout sequences safely
						if unicode.IsSpace(rune(text[x])) {
							advance(1)
							continue
						}

						// Detect regular end bounds
						if text[x] == '>' {
							advance(1)
							tagClosed = true
							break
						}

						// Support self-closing trailing forward slash bounds gracefully
						if text[x] == '/' && x+1 < len(text) && text[x+1] == '>' {
							advance(2)
							tagClosed = true
							break
						}

						// Extract attribute key definitions safely
						if unicode.IsLetter(rune(text[x])) || text[x] == '-' || text[x] == '_' {
							attrLine := line
							attrCol := col
							var attrBuf strings.Builder
							for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x])) || text[x] == '-' || text[x] == '_' || text[x] == '.') {
								attrBuf.WriteByte(text[x])
								advance(1)
							}

							tokens = append(tokens, Token{
								Type:  TokenAttrName,
								Value: attrBuf.String(),
								Line:  attrLine,
								Col:   attrCol,
							})

							// Strip optional intervening layout whitespaces safely
							for x < len(text) && unicode.IsSpace(rune(text[x])) {
								advance(1)
							}

							// Check if it's a valueless/boolean attribute or assignment
							if x < len(text) && text[x] == '=' {
								advance(1) // Consume '='

								// Strip optional post-assignment whitespaces safely
								for x < len(text) && unicode.IsSpace(rune(text[x])) {
									advance(1)
								}

								if x < len(text) && (text[x] == '"' || text[x] == '\'') {
									quoteType := text[x]
									advance(1) // Skip quote opening character

									var valBuf strings.Builder
									valLine := line
									valCol := col
									foundQuoteEnd := false

									for x < len(text) {
										if text[x] == quoteType {
											foundQuoteEnd = true
											break
										}
										valBuf.WriteByte(text[x])
										advance(1)
									}

									if !foundQuoteEnd {
										fmt.Printf("Syntax Error: Missing closing quote character matching layout strings value definitions\n")
										fmt.Printf("  --> %s:%d:%d\n", srcFile.name, valLine, valCol)
										os.Exit(0)
									}

									tokens = append(tokens, Token{
										Type:  TokenAttrValue,
										Value: valBuf.String(),
										Line:  valLine,
										Col:   valCol,
									})
									advance(1) // Skip closing quote
								} else {
									fmt.Printf("Syntax Error: Expected matching quote marks string literal wrapping attribute assignments\n")
									fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
									os.Exit(0)
								}
							} else {
								// Valueless/boolean attribute detected: record an empty string token representation cleanly
								tokens = append(tokens, Token{
									Type:  TokenAttrValue,
									Value: "",
									Line:  attrLine,
									Col:   attrCol,
								})
							}
						} else {
							fmt.Printf("Syntax Error: Unexpected symbol '%c' encountered within tag properties definition\n", text[x])
							fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
							os.Exit(0)
						}
					}

					if !tagClosed {
						fmt.Printf("Syntax Error: Missing closing delimiter '>' for opening element definition\n")
						fmt.Printf("  --> %s:%d:%d\n", srcFile.name, startLine, startCol)
						os.Exit(0)
					}
					continue
				}
			}

			// 3. FALLBACK CONTENT COLLECTOR
			textBuf.WriteByte(text[x])
			advance(1)
		}
	}

	if inRaw {
		fmt.Printf("Syntax Error: Unclosed literal target code containment tag block sequence. Missing matching '</raw>'\n")
		os.Exit(0)
	}

	flushText()
	tokens = append(tokens, Token{Type: TokenEOF, Value: "", Line: line, Col: col})
	return tokens
}
