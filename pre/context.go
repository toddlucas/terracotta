package pre

import "fmt"

type parserScope struct {
	active   bool
	branched bool
}

type parserContext struct {
	nameStack  []nameTable
	scopeStack []parserScope
	//	active     bool
	verbose bool
}

func (c *parserContext) enterNamespace() {
	c.nameStack = append(c.nameStack, *newNameTable())
	c.scopeStack = []parserScope{parserScope{true, false}}
	//c.active = true
}

func (c *parserContext) leaveNamespace() {
	// Pop
	_, c.nameStack = c.nameStack[len(c.nameStack)-1], c.nameStack[:len(c.nameStack)-1]
}

func (c *parserContext) define(name string) {
	c.nameStack[len(c.nameStack)-1].define(name)
}

func (c *parserContext) undef(name string) {
	c.nameStack[len(c.nameStack)-1].undef(name)
	//	delete(t.names, name)
}

func (c *parserContext) isDefined(name string) bool {
	for i := len(c.nameStack) - 1; i >= 0; i-- {
		if c.nameStack[i].exists(name) {
			return c.nameStack[i].defined(name)
		}
	}

	return false
}

func (c *parserContext) scope() *parserScope {
	return &c.scopeStack[len(c.scopeStack)-1]
}

func (c *parserContext) parent() *parserScope {
	return &c.scopeStack[len(c.scopeStack)-2]
}

func (c *parserContext) enterBranch() {
	active := c.scope().active
	//c.active = active
	c.scopeStack = append(c.scopeStack, parserScope{active, false})
}

func (c *parserContext) leaveBranch() {
	// Pop
	_, c.scopeStack = c.scopeStack[len(c.scopeStack)-1], c.scopeStack[:len(c.scopeStack)-1]
	//c.active = c.scope().active
}

func (c *parserContext) takeBranch(taken bool) {
	//fmt.Printf("takeBranch %t\n", taken)
	s := c.scope()
	//c.active = taken
	if taken {
		s.active = c.parent().active
	} else {
		s.active = false
	}
	s.branched = taken
}

func (c *parserContext) previousBranchTaken() bool {
	return c.scope().branched
}

func (c *parserContext) nextBranch() {
	if c.previousBranchTaken() {
		c.scope().active = false
		//c.active = false
	}
}

func (c *parserContext) evaluateExpression(e *Expression) bool {
	if c.verbose {
		fmt.Printf("evaluateExpression %d\n", e.kind)
	}
	switch e.kind {
	case ExpressionUnary:
		return c.evaluateUnaryExpression(e)
	case ExpressionBinary:
		return c.evaluateBinaryExpression(e)
	case ExpressionGroup:
		return c.evaluateGroupExpression(e)
	case ExpressionIdentifier:
		return c.evaluateIdentifierExpression(e)
	}
	fmt.Printf("ERROR: Unrecognized expression\n")
	return false
}

func (c *parserContext) evaluateUnaryExpression(e *Expression) bool {
	if c.verbose {
		fmt.Printf("evaluateUnaryExpression\n")
	}
	expression := c.evaluateExpression(e.left)
	switch e.operator {
	case TokenNot:
		return !expression
	}

	fmt.Print(tokenToString(e.operator))
	fmt.Print(e)
	// TODO: Error handling
	fmt.Printf("ERROR: Unrecognized unary expression\n")
	return false
}

func (c *parserContext) evaluateBinaryExpression(e *Expression) bool {
	if c.verbose {
		fmt.Printf("evaluateBinaryExpression\n")
	}

	left := c.evaluateExpression(e.left)
	switch e.operator {
	case TokenAnd:
		// left && right
		if !left {
			// Early out
			return false
		}
		right := c.evaluateExpression(e.right)
		return right
	case TokenOr:
		// left || right
		if left {
			// Early out
			return true
		}
		right := c.evaluateExpression(e.right)
		return right
	}

	// TODO: Error handling
	if c.verbose {
		fmt.Printf("ERROR: Unrecognized binary expression\n")
	}
	return false
}

func (c *parserContext) evaluateGroupExpression(e *Expression) bool {
	if c.verbose {
		fmt.Printf("evaluateGroupExpression\n")
	}
	return c.evaluateExpression(e.left)
}

func (c *parserContext) evaluateIdentifierExpression(e *Expression) bool {
	if c.verbose {
		fmt.Print("evaluateIdentifierExpression: ")
		result := c.isDefined(e.identifier)
		fmt.Printf("%s = %t\n", e.identifier, result)
	}
	return c.isDefined(e.identifier)
}
