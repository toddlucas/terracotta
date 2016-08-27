package pre

// Token represents a sequence of text recognized by the scanner (a terminal)
// or a parser state (non terminal).
type Token int

const (
	// TokenNone is the zero value.
	TokenNone Token = iota
	// TokenEnd is the last token the scanner will return; it corresponds to
	// the end of file or end of string.
	TokenEnd
	TokenText
	TokenLine
	TokenComment
	TokenDirective
	TokenIdentifier
	TokenAnd
	TokenOr
	TokenNot
	TokenLParen
	TokenRParen
)

func tokenToString(token Token) string {
	result := ""

	switch token {
	case TokenNone:
		result = "None"
	case TokenEnd:
		result = "End"
	case TokenLine:
		result = "Line"
	case TokenText:
		result = "Text"
	case TokenComment:
		result = "Comment"
	case TokenDirective:
		result = "Directive"
	case TokenIdentifier:
		result = "Identifier"
	case TokenAnd:
		result = "AND"
	case TokenOr:
		result = "OR"
	case TokenNot:
		result = "NOT"
	case TokenLParen:
		result = "LeftParen"
	case TokenRParen:
		result = "RightParen"
	}

	return result
}
