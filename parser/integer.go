package parser

import (
	"encoding/json"
	"strconv"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// IntegerExpr is a integer expression
type IntegerExpr struct {
	baseExpr
	Literal int64
}

// GetType returns type of integer expression
func (expr *IntegerExpr) GetType() string {
	return ValueTypeInteger
}

// IsReady returns if integer expression is ready
func (expr *IntegerExpr) IsReady(stack utils.StackInterface) (ready bool) {
	if expr.Func != nil {
		return expr.Func.isReady(stack)
	}
	return true
}

// GetValue returns value of integer expression
func (expr *IntegerExpr) GetValue(stack utils.StackInterface) (value int64, err error) {
	if expr.Func != nil {
		fValue, err := expr.Func.getValue(stack)
		if err != nil {
			return value, errors.Trace(err)
		}
		var ok bool
		if value, ok = fValue.(int64); !ok {
			return value, ExpressionValueInvalidError{
				Declaration: expr.GetDeclaration(), Type: expr.GetType()}
		}
		return value, nil
	}
	return expr.Literal, nil
}

// UnmarshalJSON sets the object from the provided JSON representation
func (expr *IntegerExpr) UnmarshalJSON(data []byte) error {
	var v int64
	err := json.Unmarshal(data, &v)
	if err == nil {
		expr.Func = nil
		expr.Literal = v
		return nil
	}

	// Cloudformation allows int values to be represented as strings
	var strValue string
	err2 := json.Unmarshal(data, &strValue)
	if err2 == nil {
		if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
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
