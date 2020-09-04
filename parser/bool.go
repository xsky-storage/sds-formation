package parser

import (
	"encoding/json"
	"strconv"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// BoolExpr is a bool expression
type BoolExpr struct {
	baseExpr
	Literal bool
}

// GetType returns type of bool expression
func (expr *BoolExpr) GetType() string {
	return ValueTypeBool
}

// IsReady returns if bool expression is ready
func (expr *BoolExpr) IsReady(stack utils.StackInterface) (ready bool) {
	if expr.Func != nil {
		return expr.Func.isReady(stack)
	}
	return true
}

// GetValue returns value of bool expression
func (expr *BoolExpr) GetValue(stack utils.StackInterface) (value bool, err error) {
	if expr.Func != nil {
		fValue, err := expr.Func.getValue(stack)
		if err != nil {
			return value, errors.Trace(err)
		}
		var ok bool
		if value, ok = fValue.(bool); !ok {
			return value, ExpressionValueInvalidError{
				Declaration: expr.GetDeclaration(), Type: expr.GetType()}
		}
		return value, nil
	}
	return expr.Literal, nil
}

// UnmarshalJSON sets the object from the provided JSON representation
func (expr *BoolExpr) UnmarshalJSON(data []byte) error {
	var v bool
	err := json.Unmarshal(data, &v)
	if err == nil {
		expr.Func = nil
		expr.Literal = v
		return nil
	}

	// Cloudformation allows bool values to be represented as the
	// strings "true" and "false"
	var strValue string
	err2 := json.Unmarshal(data, &strValue)
	if err2 == nil {
		if v, err := strconv.ParseBool(strValue); err == nil {
			expr.Func = nil
			expr.Literal = v
			return nil
		}
	}

	// Perhaps we have a serialized function call (like `{"Ref": "Foo"}`)
	// so we'll try to unmarshal it with UnmarshalFunc.
	funcCall, err3 := unmarshalFunc(expr.GetType(), data)
	if err3 == nil {
		expr.Func = funcCall
		expr.declaration = string(data)
		return nil
	} else if _, ok := errors.Cause(err3).(FunctionUnknownError); ok {
		return errors.Trace(err3)
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}
