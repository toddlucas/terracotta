package pre

import "testing"

func TestInvalidDirective(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("\r\n!\r\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectSyntaxErrorKind(t, &s, SyntaxErrorInvalidDirective)
}

func TestValidDirectiveLines(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("!define")
	scanExpectToken(t, &s, TokenDirective)
	scanExpectToken(t, &s, TokenEnd)

	s.SetText("\n!   define\r\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenDirective)
	//scanExpectToken(t, &s, TokenLine)
	scanExpectToken(t, &s, TokenEnd)

	s.SetText("\r\n! define")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenDirective)
	scanExpectToken(t, &s, TokenEnd)

	s.SetText("   ! define  \r\n\r\n")
	scanExpectToken(t, &s, TokenDirective)
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)
}

func TestLineHandling(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	//
	// *nix LF
	//

	s.SetText("\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)

	s.SetText("\n\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)

	//
	// DOS/Windows CR/LF
	//

	s.SetText("\r\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)

	s.SetText("\r\n\r\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)

	// Mac, anyone?
	s.SetText("\r\r")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)

	// Mixed LF + CR\LF
	s.SetText("\n\n\r\n\n")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)
}

func TestDirectiveText(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("!define FOO\n!if FOO\n!else\n# comment\n!endif")

	scanExpectTokenText(t, &s, TokenDirective, "define")
	scanExpectTokenText(t, &s, TokenIdentifier, "FOO")
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenDirective, "if")
	scanExpectTokenText(t, &s, TokenIdentifier, "FOO")
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenDirective, "else")
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenText, "# comment")
	scanExpectTokenText(t, &s, TokenDirective, "endif")
	scanExpectToken(t, &s, TokenEnd)
}

func TestCommentedDirective(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("\r\n#!\r\n")

	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "#!")
	scanExpectToken(t, &s, TokenEnd)
}

func TestDirectiveWithSingleComment(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("\r\n!define FOO  #  fuuu  \r\n")

	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenDirective, "define")
	scanExpectTokenText(t, &s, TokenIdentifier, "FOO")
	scanExpectToken(t, &s, TokenLine)
	// REVIEW: This is only retured if there is trailing whitespace or a comment.
	// scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)
}

func TestDirectiveWithComment(t *testing.T) {
	s := Scanner{}
	//	s.SetVerbose(true)

	s.SetText("\r\n!define FOO  /* fuuu\r\n !if FOO\r\n!endif*/ BAR\r\n\r\n")

	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenDirective, "define")
	scanExpectTokenText(t, &s, TokenIdentifier, "FOO")
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenText, " BAR")
	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectToken(t, &s, TokenEnd)
}

func TestCommentContainingDirectives(t *testing.T) {
	s := Scanner{}
	// s.SetVerbose(true)

	s.SetText("\r\nvar = ''\r\n /* !define FOO\r\n !if FOO\r\n!endif*/ var b = 1\r\n")

	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "var = ''")
	scanExpectTokenText(t, &s, TokenText, " /* !define FOO")
	scanExpectTokenText(t, &s, TokenText, " !if FOO")
	scanExpectTokenText(t, &s, TokenText, "!endif*/ var b = 1")
	scanExpectToken(t, &s, TokenEnd)
}

func TestDireciveContainingComment(t *testing.T) {
	s := Scanner{}
	// s.SetVerbose(true)

	s.SetText("\r\nvar = ''\r\n /* !define FOO*/!define BAR\r\n !if FOO\r\n")

	scanExpectTokenText(t, &s, TokenText, "")
	scanExpectTokenText(t, &s, TokenText, "var = ''")
	// TODO: We should issue a warning here, but it would require special case scanning.
	scanExpectTokenText(t, &s, TokenText, " /* !define FOO*/!define BAR")
	// scanExpectTokenText(t, &s, TokenDirective, "define")
	// scanExpectTokenText(t, &s, TokenIdentifier, "BAR")
	// scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenDirective, "if")
	scanExpectTokenText(t, &s, TokenIdentifier, "FOO")
	scanExpectToken(t, &s, TokenEnd)
}

func TestLookahead(t *testing.T) {
	s := Scanner{}
	// s.SetVerbose(true)

	s.SetText(`
!define
!if
!endif
`)

	scanExpectTokenText(t, &s, TokenText, "")
	peekExpectTokenText(t, &s, TokenDirective, DirectiveDefine)
	scanExpectTokenText(t, &s, TokenDirective, DirectiveDefine)
	scanExpectToken(t, &s, TokenLine)
	peekExpectTokenText(t, &s, TokenDirective, DirectiveIf)
	// 2nd peek
	peekExpectProcessingErrorKind(t, &s, ProcessingInvalidLookahead)
	scanExpectTokenText(t, &s, TokenDirective, DirectiveIf)
	scanExpectToken(t, &s, TokenLine)
	scanExpectTokenText(t, &s, TokenDirective, DirectiveEndif)
	scanExpectToken(t, &s, TokenEnd)
}

func TestInvalidSlash(t *testing.T) {
	s := Scanner{}
	// s.SetVerbose(true)

	s.SetText("!define/ FOO")

	scanExpectSyntaxErrorKind(t, &s, SyntaxErrorInvalidDirective)

	s.SetText("!define / FOO")

	scanExpectTokenText(t, &s, TokenDirective, DirectiveDefine)
	scanExpectSyntaxErrorKind(t, &s, SyntaxErrorInvalidDirective)
}

//
// Helpers
//

func scanExpectSyntaxErrorKind(t *testing.T, s *Scanner, expected SyntaxErrorKind) {
	_, _, err := s.Scan()
	expectSyntaxErrorKind(t, s, expected, err)
}

func peekExpectSyntaxErrorKind(t *testing.T, s *Scanner, expected SyntaxErrorKind) {
	_, _, err := s.Peek()
	expectSyntaxErrorKind(t, s, expected, err)
}

func expectSyntaxErrorKind(t *testing.T, s *Scanner, expected SyntaxErrorKind, err error) {
	if err == nil {
		t.Error("Expected syntax error")
		return
	}

	se, found := err.(SyntaxError)
	if !found {
		t.Error("Expected syntax error")
	}

	if se.Kind() != expected {
		// debug.PrintStack()
		t.Errorf("Expected syntax error kind %d but received %d", expected, se.Kind())
	}
}

func scanExpectProcessingErrorKind(t *testing.T, s *Scanner, expected ProcessingErrorKind) {
	_, _, err := s.Scan()
	expectProcessingErrorKind(t, s, expected, err)
}

func peekExpectProcessingErrorKind(t *testing.T, s *Scanner, expected ProcessingErrorKind) {
	_, _, err := s.Peek()
	expectProcessingErrorKind(t, s, expected, err)
}

func expectProcessingErrorKind(t *testing.T, s *Scanner, expected ProcessingErrorKind, err error) {
	if err == nil {
		t.Error("Expected processing error: " + err.Error())
		return
	}

	se, found := err.(ProcessingError)
	if !found {
		t.Error("Expected processing error")
	}

	if se.Kind() != expected {
		// debug.PrintStack()
		t.Errorf("Expected processing error kind %d but received %d", expected, se.Kind())
	}
}

func scanExpectToken(t *testing.T, s *Scanner, expected Token) {
	token, text, err := s.Scan()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	if expected != token {
		// debug.PrintStack()
		t.Errorf("Expected token %s but received %s with text '%s'", tokenToString(expected), tokenToString(token), text)
	}
}

func peekExpectToken(t *testing.T, s *Scanner, expected Token) {
	token, text, err := s.Peek()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	if expected != token {
		// debug.PrintStack()
		t.Errorf("Expected token %s but received %s with text '%s'", tokenToString(expected), tokenToString(token), text)
	}
}

func scanExpectTokenText(t *testing.T, s *Scanner, expectedToken Token, expectedText string) {
	token, text, err := s.Scan()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	expectTokenText(t, s, expectedToken, expectedText, token, text)
}

func peekExpectTokenText(t *testing.T, s *Scanner, expectedToken Token, expectedText string) {
	token, text, err := s.Peek()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	expectTokenText(t, s, expectedToken, expectedText, token, text)
}

func expectTokenText(t *testing.T, s *Scanner, expectedToken Token, expectedText string, token Token, text string) {
	if expectedToken != token {
		// debug.PrintStack()
		t.Errorf("Expected token %s with text '%s' but received %s with text '%s'", tokenToString(expectedToken), expectedText, tokenToString(token), text)
	} else if expectedText != text {
		// debug.PrintStack()
		t.Errorf("Expected text '%s' but received '%s' with token %s", expectedText, text, tokenToString(token))
	}
}
