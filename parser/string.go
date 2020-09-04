package parser

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// StringExpr is a string expression
type StringExpr struct {
	baseExpr
	Literal string
}

// GetType returns type of string expression
func (expr *StringExpr) GetType() string {
	return ValueTypeString
}

// IsReady returns if string expression is ready
func (expr *StringExpr) IsReady(stack utils.StackInterface) (ready bool) {
	if expr.Func != nil {
		return expr.Func.isReady(stack)
	}
	return true
}

// GetValue returns value of string expression
func (expr *StringExpr) GetValue(stack utils.StackInterface) (value string, err error) {
	if expr.Func != nil {
		fValue, err := expr.Func.getValue(stack)
		if err != nil {
			return value, errors.Trace(err)
		}
		var ok bool
		if value, ok = fValue.(string); !ok {
			return value, ExpressionValueInvalidError{
				Declaration: expr.GetDeclaration(), Type: expr.GetType()}
		}
		return value, nil
	}
	return expr.Literal, nil
}

// UnmarshalJSON sets the object from the provided JSON representation
func (expr *StringExpr) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err == nil {
		expr.Func = nil
		expr.Literal = v
		return nil
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc.
	funcCall, err2 := unmarshalFunc(expr.GetType(), data)
	if err2 == nil {
		expr.Func = funcCall
		expr.declaration = string(data)
		return nil
	} else if _, ok := errors.Cause(err2).(FunctionUnknownError); ok {
		return errors.Trace(err2)
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return errors.Trace(err)
}
