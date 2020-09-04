package parser

import "xsky.com/sds-formation/utils"

// ExprType defines expression interface
type ExprType interface {
	GetType() string
	GetDeclaration() string
	IsReady(utils.StackInterface) bool
}

// ListExprType defines list expression interface
type ListExprType interface {
	ExprType
	Select(utils.StackInterface, int) (interface{}, error)
}

type baseExpr struct {
	declaration string
	Func        Func
}

func (expr *baseExpr) GetDeclaration() string {
	return expr.declaration
}

// types of expression
const (
	ValueTypeBool        = "Bool"
	ValueTypeBoolList    = "BoolList"
	ValueTypeInteger     = "Integer"
	ValueTypeIntegerList = "IntegerList"
	ValueTypeString      = "String"
	ValueTypeStringList  = "StringList"
)
