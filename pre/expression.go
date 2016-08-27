package pre

import "fmt"

type ExpressionKind int

const (
	ExpressionNone ExpressionKind = iota
	ExpressionIdentifier
	ExpressionUnary
	ExpressionBinary
	ExpressionGroup
)

func expressionToString(kind ExpressionKind) string {
	result := ""
	switch kind {
	case ExpressionNone:
		result = "None"
	case ExpressionIdentifier:
		result = "Identifier"
	case ExpressionUnary:
		result = "Unary"
	case ExpressionBinary:
		result = "Binary"
	case ExpressionGroup:
		result = "Group"
	}
	return result
}

type Expression struct {
	kind       ExpressionKind
	operator   Token
	identifier string
	left       *Expression
	right      *Expression
}

func printExpression(level int, expression *Expression) {
	for i := 0; i < level; i++ {
		fmt.Print("  ")
	}
	fmt.Printf("[%s, %s, %s]\n", expressionToString(expression.kind), tokenToString(expression.operator), expression.identifier)
	if expression.left != nil {
		printExpression(level+1, expression.left)
	}
	if expression.right != nil {
		printExpression(level+1, expression.right)
	}
}
