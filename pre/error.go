package pre

import "fmt"

type ProcessingErrorKind int

const (
	ProcessingErrorOk ProcessingErrorKind = iota
	ProcessingInvalidState
	ProcessingInvalidLookahead
)

type ProcessingError struct {
	message string
	line    int
	column  int
	kind    ProcessingErrorKind
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("(%d): %s", e.line, e.message)
}

func (e ProcessingError) String() string {
	return e.message
}

func (e *ProcessingError) Kind() ProcessingErrorKind {
	return e.kind
}
