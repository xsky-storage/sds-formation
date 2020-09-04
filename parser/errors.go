package parser

import (
	"fmt"
)

// FunctionUnknownError defines unkonwn function error
type FunctionUnknownError struct {
	FunctionName string
}

func (e FunctionUnknownError) Error() string {
	return fmt.Sprintf("unknown function %s", e.FunctionName)
}

// ExpressionParamInsufficientError defines insufficient params of expression error
type ExpressionParamInsufficientError struct {
	ParamName string
}

func (e ExpressionParamInsufficientError) Error() string {
	return fmt.Sprintf("param %s is required", e.ParamName)
}

// ExpressionValueInvalidError defines invalid value of expression error
type ExpressionValueInvalidError struct {
	Declaration string
	Type        string
}

func (e ExpressionValueInvalidError) Error() string {
	return fmt.Sprintf("value of expression %s doesn't support expression type %s",
		e.Declaration, e.Type)
}
