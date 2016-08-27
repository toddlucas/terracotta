package pre

import (
	"bytes"
	"fmt"
	"unicode"
)

// scanState represents the internal state of the lexical analyzer.
type scanState int

const (
	scanStateInit scanState = iota
	scanStateBang
	scanStateText
	scanStateLine
	scanStateDirective
	scanStateParams
	scanStateIdentifier
	scanStateSingleComment
	scanStateMultiComment
	scanStateSlash
	scanStateHash
)

// We use a '!' since '#' is reserved for single-line comments.
const directivePrefixRune = '!'
const directivePrefixString = "!"

// Scanner is the lexical analyzer for the preprocessor.
type Scanner struct {
	buffer    readBuffer
	state     scanState
	lookahead int
	nextToken Token
	nextText  string
	nextErr   error
	line      int
	comment   bool // Is the scanner currently in a multiline comment?
	multiline int
	prev      scanState // The state previous to the comment.
	verbose   bool
}

func (s *Scanner) reset() {
	s.state = scanStateInit
	s.lookahead = 0
	s.line = 0
	s.comment = false
}

func (s *Scanner) SetFile(name string) {
	s.reset()
	s.buffer = readBuffer{}
	s.buffer.readFile(name)
}

func (s *Scanner) SetText(text string) {
	s.reset()
	s.buffer = readBuffer{}
	s.buffer.readText(text)
}

func (s *Scanner) SetVerbose(verbose bool) {
	s.verbose = verbose
}

// Line returns the current line that the scanner is evaluating.
// It can be used to supplement error messages.
func (s *Scanner) Line() int {
	return s.line
}

// Peek returns k=1 lookahead tokens.
func (s *Scanner) Peek() (Token, string, error) {
	if s.lookahead > 1 {
		return TokenNone, "", ProcessingError{"Invalid lookahead", s.line, 0, ProcessingInvalidLookahead}
	}

	if s.lookahead > 0 {
		s.lookahead++
		return s.nextToken, s.nextText, s.nextErr
	}

	s.nextToken, s.nextText, s.nextErr = s.Scan()
	s.lookahead = 2

	return s.nextToken, s.nextText, s.nextErr
}

func (s *Scanner) Push() error {
	if s.lookahead == 2 {
		s.lookahead = 1
		return nil
	}

	return ProcessingError{"Invalid push", s.line, 0, ProcessingInvalidLookahead}
}

// Scan returns the next token, if available.
func (s *Scanner) Scan() (Token, string, error) {
	if s.lookahead > 0 {
		s.lookahead = 0
		return s.nextToken, s.nextText, s.nextErr
	}

	if s.buffer.isEnd() {
		return TokenEnd, "", nil
	}

	var text bytes.Buffer
	var comment bytes.Buffer
	for {
		r := s.buffer.next()

		switch s.state {
		case scanStateInit:
			if s.verbose {
				fmt.Printf("scanStateInit %#U\n", r)
			}
			switch r {
			case ' ', '\t':
				// Ignore any whitespace before '!'
				text.WriteRune(r)
			case '\r', '\n', unicode.MaxRune:
				s.nextLine(r)
				return TokenText, text.String(), nil
			case '/':
				if s.buffer.current() == '*' {
					comment.WriteRune(r)

					s.buffer.next() // Skip *
					comment.WriteRune('*')

					s.prev = s.state
					s.state = scanStateMultiComment
					s.multiline = 0
				} else {
					comment.WriteRune(r)
				}
			case directivePrefixRune: // '!'
				s.state = scanStateBang
			default:
				s.state = scanStateText
				text.WriteRune(r)
			}

		case scanStateText:
			if s.verbose {
				fmt.Printf("scanStateText %#U\n", r)
			}
			switch r {
			case '\r', '\n', unicode.MaxRune:
				s.nextLine(r)
				s.state = scanStateInit
				return TokenText, text.String(), nil
			case '/':
				if s.buffer.current() == '*' {
					text.WriteRune(r)

					s.buffer.next() // Skip *
					text.WriteRune('*')

					s.prev = s.state
					s.state = scanStateMultiComment
					s.multiline = 0
				} else {
					text.WriteRune(r)
				}
			default:
				s.state = scanStateText
				text.WriteRune(r)
			}

		case scanStateBang:
			if s.verbose {
				fmt.Printf("scanStateBang %#U\n", r)
			}
			switch {
			case r == ' ' || r == '\t':
				// Skip any whitespace after '!'
			case r == '_' || unicode.IsLetter(r):
				text.Reset()
				text.WriteRune(r)
				s.state = scanStateDirective
			default:
				return TokenNone, text.String(), SyntaxError{"Invalid directive", s.line, 0, SyntaxErrorInvalidDirective}
			}

		case scanStateDirective:
			if s.verbose {
				fmt.Printf("scanStateDirective %#U\n", r)
			}
			switch {
			case r == '_' || unicode.IsLetter(r):
				text.WriteRune(r)
			case r == '\r' || r == '\n' || r == unicode.MaxRune:
				s.nextLine(r)
				// if r != unicode.MaxRune {
				// 	s.buffer.push()
				// }
				s.state = scanStateLine
				// s.nextLine(r)
				// s.state = scanStateInit
				return TokenDirective, text.String(), nil
			case r == '#':
				// Any single-line comments after a directive get eaten.
				s.prev = s.state
				s.state = scanStateSingleComment
			case r == '/':
				// Multiline comments also get eaten.
				if s.buffer.current() == '*' {
					s.buffer.next() // Skip *
					s.prev = s.state
					s.state = scanStateMultiComment
					s.multiline = 0
				} else {
					return TokenNone, text.String(), SyntaxError{"Invalid / in directive", s.line, 0, SyntaxErrorInvalidDirective}
				}
			default:
				s.state = scanStateParams
				s.buffer.push()
				return TokenDirective, text.String(), nil
			}

		case scanStateParams:
			if s.verbose {
				fmt.Printf("scanStateParams %#U\n", r)
			}
			switch {
			case r == ' ' || r == '\t':
				// Ignore interstitial whitespace
			case r == '\r' || r == '\n' || r == unicode.MaxRune:
				s.nextLine(r)
				// if r != unicode.MaxRune {
				// 	s.buffer.push()
				// }
				s.state = scanStateLine
				// s.nextLine(r)
				// text.Reset()
				// s.state = scanStateInit
			case r == '_' || unicode.IsLetter(r):
				s.state = scanStateIdentifier
				text.WriteRune(r)
			case r == '&':
				if s.buffer.current() == '&' {
					s.buffer.next() // Skip second &
					return TokenAnd, "", nil
				}
				return TokenNone, text.String(), SyntaxError{"Invalid operator: AND is &&", s.line, 0, SyntaxErrorInvalidOperator}
			case r == '|':
				if s.buffer.current() == '|' {
					s.buffer.next() // Skip second |
					return TokenOr, "", nil
				}
				return TokenNone, text.String(), SyntaxError{"Invalid operator: OR is ||", s.line, 0, SyntaxErrorInvalidOperator}
			case r == '!':
				return TokenNot, "", nil
			case r == '(':
				return TokenLParen, "", nil
			case r == ')':
				return TokenRParen, "", nil
			case r == '#':
				// Any single-line comments after a directive get eaten.
				text.WriteRune(r)
				s.prev = s.state
				s.state = scanStateSingleComment
			case r == '/':
				// Multiline comments also get eaten.
				if s.buffer.current() == '*' {
					s.buffer.next() // Skip *
					s.prev = s.state
					s.state = scanStateMultiComment
					s.multiline = 0
				} else {
					return TokenNone, text.String(), SyntaxError{"Invalid / following directive", s.line, 0, SyntaxErrorInvalidDirective}
				}
			default:
				return TokenNone, text.String(), SyntaxError{"Invalid directive parameters", s.line, 0, SyntaxErrorInvalidParameters}
			}

		case scanStateIdentifier:
			if s.verbose {
				fmt.Printf("scanStateIdentifier %#U\n", r)
			}
			switch {
			case r == '\r' || r == '\n' || r == unicode.MaxRune:
				s.nextLine(r)
				// if r != unicode.MaxRune {
				// 	s.buffer.push()
				// }
				s.state = scanStateLine
				// s.nextLine(r)
				// s.state = scanStateInit
				return TokenIdentifier, text.String(), nil
			case r == '_' || unicode.IsLetter(r):
				text.WriteRune(r)
			case r == ' ' || r == '\t':
				s.state = scanStateParams
				return TokenIdentifier, text.String(), nil
			default:
				s.buffer.push()
				s.state = scanStateParams
				return TokenIdentifier, text.String(), nil
			}

		case scanStateLine:
			if s.verbose {
				fmt.Printf("scanStateLine\n")
			}

			// We're emitting a line, so ignore the current rune.
			if r != unicode.MaxRune {
				s.buffer.push()
			}

			// switch {
			// case r == '\r' || r == '\n' || r == unicode.MaxRune:
			// 	s.nextLine(r)
			s.state = scanStateInit
			return TokenLine, "", nil
			// default:
			// 	return TokenNone, text.String(), ProcessingError{"Invalid state", s.line, 0, ProcessingInvalidState}
			// }

		case scanStateSingleComment:
			if s.verbose {
				fmt.Printf("scanStateSingleComment %#U\n", r)
			}
			switch {
			case r == '\r' || r == '\n' || r == unicode.MaxRune:
				s.buffer.push()
				s.state = s.prev
			default:
				text.WriteRune(r)
			}

		case scanStateMultiComment:
			if s.verbose {
				fmt.Printf("scanStateMultiComment %#U\n", r)
			}
			switch r {
			case '*':
				if s.buffer.current() == '/' {
					s.buffer.next() // Skip /

					if s.prev == scanStateText {
						text.WriteRune(r)
						text.WriteRune('/')
					} else if s.multiline > 0 {
						// If we were in a preprocessor state and the comment
						// was on the same line, then we do nothing. If the
						// comment spanned multiple lines, then we simulate
						// an EOL.
						s.prev = scanStateLine
					}

					s.multiline = 0
					s.state = s.prev
				} else {
					if s.prev == scanStateInit {
						comment.WriteRune(r)
					} else if s.prev == scanStateText {
						text.WriteRune(r)
					}
				}
			case '\r', '\n', unicode.MaxRune:
				s.nextLine(r)
				s.multiline++

				if s.prev == scanStateInit {
					// If we started in Init, but this is now multiline,
					// we treat it as Text. Making this distinction allows
					// us to resume Init processing for single line variants.
					text.WriteString(comment.String())
					s.prev = scanStateText
				}

				if s.prev == scanStateText {
					return TokenText, text.String(), nil
					// } else {
					// 	if r != unicode.MaxRune {
					// 		s.buffer.push()
					// 	}
					// 	s.state = scanStateLine
				}
			default:
				if s.prev == scanStateInit {
					comment.WriteRune(r)
				} else if s.prev == scanStateText {
					text.WriteRune(r)
				}
			}
		}
	}
}

// chomp will eat any remaining end of line characters.
// This is mainly useful on Windows, which uses CR\LF.
// NOTE: Should not eat any following lines.
func (s *Scanner) chomp(current rune) {
	if current == '\r' {
		r := s.buffer.current()
		if r == '\n' {
			s.buffer.next()
		}
	}
}

// nextLine handles common end of line processing.
func (s *Scanner) nextLine(current rune) {
	s.chomp(current)
	s.line++
	if s.verbose {
		fmt.Println("LINE")
	}
}
