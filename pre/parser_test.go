package pre

import "testing"

func TestParseSimpleExpression(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	//	p.SetText("\n!define FOO\n!if !(!FOO && BAR)\nvar a = 1")
	p.SetText("\n!define FOO\n!if !(!FOO && !BAR)\nvar a = 1")

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectDirective(t, &p)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "var a = 1", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseExpressionVariations(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(false, false, true)

	//	p.SetText("\n!define FOO\n!if !(!FOO && BAR)\nvar a = 1")
	p.SetText(`
!if false /*
*
*/
`)

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "", false)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseBasicConditions(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	//	p.SetText("\n!define FOO\n!if !(!FOO && BAR)\nvar a = 1")
	p.SetText("\n!define FOO\n!if BAR\nbad\n!elif FOO\ngood\n!else\nbad\n!endif\ngood\n")

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectDirective(t, &p)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "bad", false)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "good", true)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "bad", false)
	parseExpectDirective(t, &p)
	parseExpectText(t, &p, "good", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseExtendedConditions(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	p.SetText(`
	/*
	* Multiline comment
	*/
!define FOO
!define BAR /*
* Masked multiline comment
*/
0 good
!if BAZ
	1 bad
!elif FOO
	2 good
	!if BAR
		3 good
	!else
		4 bad
	!endif
	5 good
	!if BAZ
		6 bad
	!else
		7 good
	!endif
	8 good
!elif BAR
	9 bad
	!if FOO
		10 bad
	!else
		11 bad
	!endif
	12 bad
!else
	13 bad
!endif
14 good`)

	// TODO: Test with no enter
	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectText(t, &p, "\t/*", true)
	parseExpectText(t, &p, "\t* Multiline comment", true)
	parseExpectText(t, &p, "\t*/", true)
	//  !define FOO
	parseExpectDirective(t, &p)
	//  !define BAR
	parseExpectDirective(t, &p)

	// REVIEW: There is an extra line emitted at the end of a multiline
	// comment that trails a directive. This is because the multiline
	// comment itself causes a line token to be emitted. Then the trailing
	// newline emits a second line. If we wanted formatting to behave as
	// expected, we might eat a trailing newline.
	parseExpectText(t, &p, "", true)
	//  0 good
	parseExpectText(t, &p, "0 good", true)
	//  !if BAZ
	parseExpectDirective(t, &p)
	//  	1 bad
	parseExpectText(t, &p, "\t1 bad", false)
	//  !elif FOO
	parseExpectDirective(t, &p)
	//  	2 good
	parseExpectText(t, &p, "\t2 good", true)
	//  	!if BAR
	parseExpectDirective(t, &p)
	//  		3 good
	parseExpectText(t, &p, "\t\t3 good", true)
	//  	!else
	parseExpectDirective(t, &p)
	//  		4 bad
	parseExpectText(t, &p, "\t\t4 bad", false)
	//  	!endif
	parseExpectDirective(t, &p)
	//  	5 good
	parseExpectText(t, &p, "\t5 good", true)
	//  	!if BAZ
	parseExpectDirective(t, &p)
	//  		6 bad
	parseExpectText(t, &p, "\t\t6 bad", false)
	//  	!else
	parseExpectDirective(t, &p)
	//  		7 good
	parseExpectText(t, &p, "\t\t7 good", true)
	//  	!endif
	parseExpectDirective(t, &p)
	//  	8 good
	parseExpectText(t, &p, "\t8 good", true)
	//  !elif BAR
	parseExpectDirective(t, &p)
	//  	9 bad
	parseExpectText(t, &p, "\t9 bad", false)
	//  	!if FOO
	parseExpectDirective(t, &p)
	//  		10 bad
	parseExpectText(t, &p, "\t\t10 bad", false)
	//  	!else
	parseExpectDirective(t, &p)
	//  		11 bad
	parseExpectText(t, &p, "\t\t11 bad", false)
	//  	!endif
	parseExpectDirective(t, &p)
	//  	12 bad
	parseExpectText(t, &p, "\t12 bad", false)
	//  !else
	parseExpectDirective(t, &p)
	//  	13 bad
	parseExpectText(t, &p, "\t13 bad", false)
	//  !endif
	parseExpectDirective(t, &p)
	//  14 good
	parseExpectText(t, &p, "14 good", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseEndOfFileMultilineComment(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	// Ensure that unterminated comments at the end of the file are emitted.
	p.SetText("\n/* Unterminated comment\n*\n")

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectText(t, &p, "/* Unterminated comment", true)
	parseExpectText(t, &p, "*", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseSlashRegression(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	// Ensure that slash in text is unaltered by comment handling.
	p.SetText("\nsource = \"./vpc\"")

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectText(t, &p, "source = \"./vpc\"", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func TestParseMultilineCommentRegression(t *testing.T) {
	p := Parser{}
	// p.SetVerbose(true, true, true)

	// Ensure that multiline comments that don't span lines are emitted.
	p.SetText("\n/* Keep me */\n\n  /* And me */ \n/\n")

	p.Enter()
	parseExpectText(t, &p, "", true)
	parseExpectText(t, &p, "/* Keep me */", true)
	parseExpectText(t, &p, "", true)
	parseExpectText(t, &p, "  /* And me */ ", true)
	parseExpectText(t, &p, "/", true)
	parseExpectEnd(t, &p)
	p.Leave()
}

func parseExpectDirective(t *testing.T, p *Parser) {
	item, err := p.ParseLine()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	if item.kind != ParseItemDirective {
		// debug.PrintStack()
		t.Errorf("Expected directive received %d on line %d with text '%s'", item.kind, item.line, item.text)
	}
}

func parseExpectText(t *testing.T, p *Parser, text string, active bool) {
	item, err := p.ParseLine()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	if item.kind != ParseItemText {
		// debug.PrintStack()
		t.Errorf("Expected text '%s' but received %d on line %d with text '%s'", text, item.kind, item.line, item.text)
	}

	if item.text != text {
		// debug.PrintStack()
		t.Errorf("Expected text '%s' but received text '%s' on line %d", text, item.text, item.line)
	}

	if item.active != active {
		// debug.PrintStack()
		t.Errorf("Expected text '%s' active = %t on line %d", text, active, item.line)
	}
}

func parseExpectEnd(t *testing.T, p *Parser) {
	item, err := p.ParseLine()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
		return
	}

	if item.kind != ParseItemEnd {
		// debug.PrintStack()
		t.Errorf("Expected end received %d on line %d with text '%s'", item.kind, item.line, item.text)
	}
}
