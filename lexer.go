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
			// ==========================================
			// STATE 1: RAW SCOPE MODE (<raw>...</raw>)
			// ==========================================
			if x+5 < len(text) && text[x:x+6] == "</raw>" {
				flushText() // Emit everything collected inside raw block up to this tag

				tokens = append(tokens, Token{
					Type:  TokenRawClose,
					Value: "raw",
					Line:  line,
					Col:   col,
				})
				advance(6) // Skip past "</raw>"
				inRaw = false
			} else {
				// Blindly capture character without processing logic
				textBuf.WriteByte(text[x])
				advance(1)
			}
		} else {
			// ==========================================
			// STATE 2: NORMAL HTML / COMPONENT MODE
			// ==========================================

			// 1. VARIABLE INTERPOLATION DETECTOR: {{ ... }}
			if x+1 < len(text) && text[x] == '{' && text[x+1] == '{' {
				flushText()
				startLine, startCol := line, col
				advance(2) // Skip "{{"

				var varBuf strings.Builder
				varClosed := false
				for x < len(text) {
					if x+1 < len(text) && text[x] == '}' && text[x+1] == '}' {
						advance(2) // Skip "}}"
						varClosed = true
						break
					}
					varBuf.WriteByte(text[x])
					advance(1)
				}

				if !varClosed {
					fmt.Printf("Syntax Error: Unclosed variable interpolation block '}}'\n")
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

			// 2. TAG DETECTOR: <
			if text[x] == '<' {
				// Look ahead to check if this is an actual tag sequence or just a raw '<' character
				isTag := false
				if x+1 < len(text) {
					nextChar := text[x+1]
					if nextChar == '/' || nextChar == '!' || unicode.IsLetter(rune(nextChar)) {
						isTag = true
					}
				}

				if isTag {
					flushText()
					startLine, startCol := line, col

					// Scenario A: DOCTYPE declaration handling
					if x+14 < len(text) && strings.ToLower(text[x:x+15]) == "<!doctype html>" {
						tokens = append(tokens, Token{
							Type:  TokenDocTypeDecl,
							Value: text[x : x+15],
							Line:  startLine,
							Col:   startCol,
						})
						advance(15)
						continue
					}

					// Scenario B: Special raw-entry block tracker: <raw>
					if x+4 < len(text) && text[x:x+5] == "<raw>" {
						tokens = append(tokens, Token{
							Type:  TokenRawOpen,
							Value: "raw",
							Line:  startLine,
							Col:   startCol,
						})
						advance(5)
						inRaw = true
						continue
					}

					// Scenario C: Closing tag structural element: </TagName>
					if x+1 < len(text) && text[x+1] == '/' {
						tokens = append(tokens, Token{
							Type:  TokenTagClose,
							Value: "",
							Line:  startLine,
							Col:   startCol,
						})
						advance(2) // Skip "</"

						// Isolate the element name string
						var nameBuf strings.Builder
						for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x]))) {
							nameBuf.WriteByte(text[x])
							advance(1)
						}

						if nameBuf.Len() == 0 {
							fmt.Printf("Syntax Error: Missing tag name after '</'\n")
							fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
							os.Exit(0)
						}

						tokens = append(tokens, Token{
							Type:  TokenTagName,
							Value: nameBuf.String(),
							Line:  startLine,
							Col:   startCol,
						})

						// Clear optional trailing whitespaces inside tag structure
						for x < len(text) && (text[x] == ' ' || text[x] == '\t' || text[x] == '\n' || text[x] == '\r') {
							advance(1)
						}

						if x < len(text) && text[x] == '>' {
							advance(1)
						} else {
							fmt.Printf("Syntax Error: Expected closing '>' for tag closing declaration\n")
							fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
							os.Exit(0)
						}
						continue
					}

					// Scenario D: Regular Opening Tag Structure: <TagName attr="...">
					tokens = append(tokens, Token{
						Type:  TokenTagOpen,
						Value: "",
						Line:  startLine,
						Col:   startCol,
					})
					advance(1) // Skip "<"

					// Isolate Opening Tag Name
					var nameBuf strings.Builder
					for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x]))) {
						nameBuf.WriteByte(text[x])
						advance(1)
					}

					if nameBuf.Len() == 0 {
						fmt.Printf("Syntax Error: Expected valid tag name after '<'\n")
						fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
						os.Exit(0)
					}

					tokens = append(tokens, Token{
						Type:  TokenTagName,
						Value: nameBuf.String(),
						Line:  startLine,
						Col:   startCol,
					})

					// Process loop for properties/attributes inside opening tag structure
					tagClosed := false
					for x < len(text) {
						// CRITICAL SPEC REQUIREMENT: Check for invalid self-closing sequences "/>"
						if text[x] == '/' && x+1 < len(text) && text[x+1] == '>' {
							fmt.Printf("Syntax Error: Self-closing tags standard '/>' are forbidden in Lemon Markup Spec v0.2.0\n")
							fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
							os.Exit(0)
						}

						if text[x] == '>' {
							advance(1)
							tagClosed = true
							break
						}

						if text[x] == ' ' || text[x] == '\t' || text[x] == '\n' || text[x] == '\r' {
							advance(1)
							continue
						}

						// Extract attribute/prop key text strings
						var attrNameBuf strings.Builder
						attrLine, attrCol := line, col
						for x < len(text) && (unicode.IsLetter(rune(text[x])) || unicode.IsDigit(rune(text[x])) || text[x] == '-' || text[x] == '_') {
							attrNameBuf.WriteByte(text[x])
							advance(1)
						}

						if attrNameBuf.Len() > 0 {
							tokens = append(tokens, Token{
								Type:  TokenAttrName,
								Value: attrNameBuf.String(),
								Line:  attrLine,
								Col:   attrCol,
							})

							// Skip whitespace trailing the attribute name
							for x < len(text) && (text[x] == ' ' || text[x] == '\t' || text[x] == '\n' || text[x] == '\r') {
								advance(1)
							}

							// Handle assignment expressions
							if x < len(text) && text[x] == '=' {
								advance(1) // Skip "="

								// Skip whitespace ahead of value quotes
								for x < len(text) && (text[x] == ' ' || text[x] == '\t' || text[x] == '\n' || text[x] == '\r') {
									advance(1)
								}

								// Capture quoted values
								if x < len(text) && (text[x] == '"' || text[x] == '\'') {
									quoteType := text[x]
									valLine, valCol := line, col
									advance(1) // Skip opening quote

									var valBuf strings.Builder
									valueClosed := false
									for x < len(text) {
										if text[x] == quoteType {
											valueClosed = true
											break
										}
										// Standard escape character logic inside property value declarations
										if text[x] == '\\' && x+1 < len(text) && text[x+1] == quoteType {
											advance(1)
											valBuf.WriteByte(text[x])
											advance(1)
										} else {
											valBuf.WriteByte(text[x])
											advance(1)
										}
									}

									if !valueClosed {
										fmt.Printf("Syntax Error: Unterminated attribute value string context. Expected closing character (%c)\n", quoteType)
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
							}
						} else {
							// Found an unidentifiable garbage sequence inside the HTML tag block
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

			// 3. FALLBACK CONTENT COLLECTOR (Normal Text and structural spaces)
			textBuf.WriteByte(text[x])
			advance(1)
		}
	}

	if inRaw {
		fmt.Printf("Syntax Error: Unclosed literal target code segment wrapper block. Missing matching '</raw>'\n")
		fmt.Printf("  --> %s:%d:%d\n", srcFile.name, line, col)
		os.Exit(0)
	}

	flushText()
	tokens = append(tokens, Token{Type: TokenEOF, Line: line, Col: col})
	return tokens
}
