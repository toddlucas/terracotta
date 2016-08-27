package pre

import "fmt"

type SyntaxErrorKind int

const (
	SyntaxErrorOk SyntaxErrorKind = iota
	SyntaxErrorInvalidDirective
	SyntaxErrorInvalidOperator
	SyntaxErrorInvalidParameters
	SyntaxErrorInvalidExpression
	SyntaxErrorNoConditionals
	SyntaxErrorUnrecognizedDirective
	SyntaxErrorExpectedIdentifier
	SyntaxErrorPredefinedSymbol
)

type SyntaxError struct {
	message string
	line    int
	column  int
	kind    SyntaxErrorKind
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("(%d): %s", e.line, e.message)
}

func (e SyntaxError) String() string {
	return e.message
}

func (e *SyntaxError) Kind() SyntaxErrorKind {
	return e.kind
}
