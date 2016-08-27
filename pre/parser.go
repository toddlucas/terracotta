package pre

import "fmt"

const (
	DirectiveDefine = "define"
	DirectiveUndef  = "undef"
	DirectiveIf     = "if"
	DirectiveElif   = "elif"
	DirectiveElse   = "else"
	DirectiveEndif  = "endif"
)

type Parser struct {
	scanner Scanner
	context parserContext
	verbose bool
}

func (p *Parser) SetFile(name string) {
	p.scanner.SetFile(name)
}

func (p *Parser) SetText(text string) {
	p.scanner.SetText(text)
}

func (p *Parser) SetVerbose(scanner bool, parser bool, context bool) {
	p.scanner.SetVerbose(scanner)
	p.verbose = parser
	p.context.verbose = context
}

type ParseItemKind int

const (
	ParseItemNone ParseItemKind = iota
	ParseItemDirective
	ParseItemText
	ParseItemEnd
)

type ParseItem struct {
	kind   ParseItemKind
	text   string
	line   int
	active bool
}

type LineCallback func(line string)

func (p *Parser) Parse(callback LineCallback) error {
	p.Enter()
	for {
		item, err := p.ParseLine()
		if err != nil {
			return err
		}

		if item.kind == ParseItemText {
			if item.active {
				callback(item.text)
			}
		} else if item.kind == ParseItemEnd {
			break
		}
	}
	p.Leave()
	return nil
}

func (p *Parser) ParseDefines() error {
	for {
		item, err := p.ParseLine()
		if err != nil {
			return err
		}

		if item.kind == ParseItemDirective {
			switch item.text {
			case DirectiveDefine, DirectiveUndef:
				// Ok
			default:
				return SyntaxError{"Directives file may not contain conditionals", p.scanner.Line(), 0, SyntaxErrorNoConditionals}
			}
		} else if item.kind == ParseItemEnd {
			break
		} else {
			// Ignore
		}
	}
	return nil
}

func (p *Parser) Enter() {
	p.context.enterNamespace()
}

func (p *Parser) Leave() {
	p.context.enterNamespace()
}

func (p *Parser) Define(symbol string) error {
	if symbol == "true" {
		return SyntaxError{"true is a predefined symbol", p.scanner.Line(), 0, SyntaxErrorPredefinedSymbol}
	}
	if symbol == "false" {
		return SyntaxError{"false is a predefined symbol", p.scanner.Line(), 0, SyntaxErrorPredefinedSymbol}
	}
	p.context.define(symbol)
	return nil
}

func (p *Parser) Undef(symbol string) error {
	if symbol == "true" {
		return SyntaxError{"true is a predefined symbol", p.scanner.Line(), 0, SyntaxErrorPredefinedSymbol}
	}
	if symbol == "false" {
		return SyntaxError{"false is a predefined symbol", p.scanner.Line(), 0, SyntaxErrorPredefinedSymbol}
	}
	p.context.undef(symbol)
	return nil
}

func (p *Parser) ParseLine() (ParseItem, error) {
	if p.verbose {
		fmt.Printf("ParseLine\n")
	}

	token, text, err := p.scanner.Scan()
	if err != nil {
		return ParseItem{}, err
	}

	switch token {
	case TokenDirective:
		err = p.parseDirective(text)
		if err != nil {
			return ParseItem{}, err
		}
		return ParseItem{ParseItemDirective, text, p.scanner.Line(), p.IsActive()}, nil
	case TokenText:
		return ParseItem{ParseItemText, text, p.scanner.Line(), p.IsActive()}, nil
	case TokenEnd:
		return ParseItem{ParseItemEnd, "", p.scanner.Line(), false}, nil
	}

	// fmt.Print(tokenToString(token))
	// REVIEW: Kind
	return ParseItem{}, SyntaxError{"Unexpected line expression", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
}

func (p *Parser) IsActive() bool {
	return p.context.scope().active
}

func (p *Parser) parseDirective(directive string) error {
	if p.verbose {
		fmt.Printf("parseDirective %s\n", directive)
	}

	var result error
	switch directive {
	case DirectiveDefine: // "define"
		result = p.parseDefine()
	case DirectiveUndef: // "undef"
		result = p.parseUndef()
	case DirectiveIf: // "if"
		result = p.parseIf()
	case DirectiveElif: // "elif"
		result = p.parseElIf()
	case DirectiveElse: // "else"
		result = p.parseElse()
	case DirectiveEndif: // "endif"
		result = p.parseEndIf()
	default:
		message := fmt.Sprintf("Unrecognized directive %s\n", directive)
		result = SyntaxError{message, p.scanner.Line(), 0, SyntaxErrorUnrecognizedDirective}
	}

	return result
}

func (p *Parser) parseDefine() error {
	if p.verbose {
		fmt.Printf("parseDefine\n")
	}
	token, text, err := p.scanner.Scan()
	if err != nil {
		return err
	}

	switch token {
	case TokenIdentifier:
		_, err = p.parseSymbolDefine(text)
	default:
		return SyntaxError{"!define expected an identifier", p.scanner.Line(), 0, SyntaxErrorExpectedIdentifier}
	}

	return err
}

func (p *Parser) parseUndef() error {
	if p.verbose {
		fmt.Printf("parseUndef\n")
	}
	token, text, err := p.scanner.Scan()
	if err != nil {
		return err
	}

	switch token {
	case TokenIdentifier:
		_, err = p.parseSymbolUndef(text)
	default:
		return SyntaxError{"!undef expected an identifier", p.scanner.Line(), 0, SyntaxErrorExpectedIdentifier}
	}

	return err
}

func (p *Parser) parseIf() error {
	if p.verbose {
		fmt.Printf("parseIf\n")
	}

	expression, err := p.parseExpression()
	if err != nil {
		// return SyntaxError{"!if parse error " + err.Error(), p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
		return err
	}

	if p.context.verbose {
		printExpression(1, &expression)
	}

	// We don't expect anything else after the expression.
	err = p.expectDirectiveEnd(DirectiveIf)
	if err != nil {
		return err
	}

	p.context.enterBranch()

	result := p.context.evaluateExpression(&expression)
	if p.context.verbose {
		fmt.Printf("!if %t\n", result)
	}
	p.context.takeBranch(result)

	return nil
}

func (p *Parser) parseElIf() error {
	if p.verbose {
		fmt.Printf("parseElIf\n")
	}

	expression, err := p.parseExpression()
	if err != nil {
		// return SyntaxError{"!elif parse error " + err.Error(), p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
		return err
	}

	if p.context.verbose {
		printExpression(1, &expression)
	}

	// We don't expect anything else after the expression.
	err = p.expectDirectiveEnd(DirectiveElif)
	if err != nil {
		return err
	}

	p.context.nextBranch()

	if !p.context.previousBranchTaken() {
		result := p.context.evaluateExpression(&expression)
		if p.context.verbose {
			fmt.Printf("!elif %t\n", result)
		}

		p.context.takeBranch(result)
	} else {
		if p.context.verbose {
			fmt.Printf("!elif not evaluated\n")
		}
	}

	return nil
}

func (p *Parser) parseElse() error {
	if p.verbose {
		fmt.Printf("parseElse\n")
	}

	err := p.expectDirectiveEnd(DirectiveElse)
	if err != nil {
		return err
	}

	p.context.nextBranch()

	if !p.context.previousBranchTaken() {
		if p.context.verbose {
			fmt.Printf("!else\n")
		}

		p.context.takeBranch(true)
	} else {
		if p.context.verbose {
			fmt.Printf("!else not taken\n")
		}
	}

	return nil
}

func (p *Parser) parseEndIf() error {
	if p.verbose {
		fmt.Printf("parseEndIf\n")
	}

	err := p.expectDirectiveEnd(DirectiveEndif)
	if err != nil {
		return err
	}

	p.context.leaveBranch()

	return nil
}

func (p *Parser) expectDirectiveEnd(directive string) error {
	token, _, err := p.scanner.Peek()
	if err != nil {
		return err
	}

	switch token {
	case TokenLine, TokenEnd:
		p.scanner.Scan() // Eat the EOL/EOF
	default:
		message := fmt.Sprintf("!%s contains an error", directive)
		return SyntaxError{message, p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	return nil
}

func (p *Parser) parseExpression() (Expression, error) {
	if p.verbose {
		fmt.Printf("parseConditionalExpression\n")
	}

	expression, err := p.parseTerm()
	if err != nil {
		return Expression{}, err
	}

	token, _, err := p.scanner.Peek()
	if err != nil {
		return Expression{}, err
	}

	for token == TokenOr {
		p.scanner.Scan()

		left := expression

		right, err := p.parseTerm()
		if err != nil {
			return Expression{}, err
		}

		// Wrap in binary expression
		expression = Expression{ExpressionBinary, TokenOr, "", &left, &right}

		token, _, err = p.scanner.Peek()
		if err != nil {
			return Expression{}, err
		}
	}

	err = p.scanner.Push()
	if err != nil {
		return Expression{}, err
	}

	return expression, nil
}

func (p *Parser) parseTerm() (Expression, error) {
	if p.verbose {
		fmt.Printf("parseConditionalExpression\n")
	}

	expression, err := p.parseFactor()
	if err != nil {
		return Expression{}, err
	}

	token, _, err := p.scanner.Peek()
	if err != nil {
		return Expression{}, err
	}

	for token == TokenAnd {
		p.scanner.Scan()

		left := expression

		right, err := p.parseFactor()
		if err != nil {
			return Expression{}, err
		}

		// Wrap in binary expression
		expression = Expression{ExpressionBinary, TokenAnd, "", &left, &right}

		token, _, err = p.scanner.Peek()
		if err != nil {
			return Expression{}, err
		}
	}

	err = p.scanner.Push()
	if err != nil {
		return Expression{}, err
	}

	return expression, nil
}

func (p *Parser) parseFactor() (Expression, error) {
	if p.verbose {
		fmt.Printf("parseFactor\n")
	}

	token, text, err := p.scanner.Scan()
	if err != nil {
		return Expression{}, err
	}

	var result Expression
	// var err error
	switch token {
	case TokenIdentifier:
		// result, err = p.parseIdentifier(text)
		result = Expression{ExpressionIdentifier, TokenNone, text, nil, nil}
	case TokenLParen:
		result, err = p.parseGroup()
	case TokenNot:
		result, err = p.parseNot()
	default:
		// fmt.Printf("Unexpected token in expression: %s\n", text)
		return Expression{}, SyntaxError{"Unexpected token", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	return result, err
}

func (p *Parser) parseGroup() (Expression, error) {
	if p.verbose {
		fmt.Printf("parseGroup\n")
	}

	token, _, err := p.scanner.Peek()
	if err != nil {
		return Expression{}, err
	}

	switch token {
	case TokenRParen:
		return Expression{}, SyntaxError{"Expression group cannot be empty", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	err = p.scanner.Push()
	if err != nil {
		return Expression{}, err
	}

	expression, err := p.parseExpression()
	if err != nil {
		return Expression{}, err
	}

	// We expect a closing paren here.
	token, _, err = p.scanner.Peek()
	if err != nil {
		return Expression{}, err
	}

	switch token {
	case TokenRParen:
		p.scanner.Scan() // Eat the )
	default:
		return Expression{}, SyntaxError{"Expression group requires a closing parenthesis", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	// Wrap in grouping expression
	result := Expression{ExpressionGroup, TokenNone, "", &expression, nil}

	return result, nil
}

func (p *Parser) parseNot() (Expression, error) {
	if p.verbose {
		fmt.Printf("parseNot\n")
	}

	expression, err := p.parseFactor()
	if err != nil {
		return Expression{}, err
	}

	// Wrap in unary expression
	result := Expression{ExpressionUnary, TokenNot, "", &expression, nil}

	return result, nil
}

func (p *Parser) parseSymbolDefine(text string) (Expression, error) {
	if p.verbose {
		fmt.Printf("parseSymbolDefine %s\n", text)
	}

	token, _, err := p.scanner.Scan()
	if err != nil {
		return Expression{}, err
	}

	expression := Expression{ExpressionIdentifier, TokenNone, text, nil, nil}

	var result Expression
	switch token {
	case TokenLine, TokenEnd:
		err = p.Define(text)
		result = expression
	default:
		return Expression{}, SyntaxError{"define expects a symbol", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	return result, err
}

func (p *Parser) parseSymbolUndef(text string) (Expression, error) {
	if p.verbose {
		fmt.Printf("parseSymbolUndef %s\n", text)
	}

	token, _, err := p.scanner.Scan()
	if err != nil {
		return Expression{}, err
	}

	expression := Expression{ExpressionIdentifier, TokenNone, text, nil, nil}

	var result Expression
	switch token {
	case TokenLine, TokenEnd:
		err = p.Undef(text)
		result = expression
	default:
		return Expression{}, SyntaxError{"undef expects a symbol", p.scanner.Line(), 0, SyntaxErrorInvalidExpression}
	}

	return result, err
}
