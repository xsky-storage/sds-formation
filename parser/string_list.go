package parser

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// StringListExpr is a string list expression
type StringListExpr struct {
	baseExpr
	Literal []*StringExpr
}

// GetType returns type of string list expression
func (expr *StringListExpr) GetType() string {
	return ValueTypeStringList
}

// IsReady returns if string list expression is ready
func (expr *StringListExpr) IsReady(stack utils.StackInterface) (ready bool) {
	if expr.Func != nil {
		return expr.Func.isReady(stack)
	}
	for _, subExpr := range expr.Literal {
		if !subExpr.IsReady(stack) {
			return false
		}
	}
	return true
}

// GetValue returns value of string list expression
func (expr *StringListExpr) GetValue(stack utils.StackInterface) (value []string, err error) {
	if expr.Func != nil {
		fValue, err := expr.Func.getValue(stack)
		if err != nil {
			return value, errors.Trace(err)
		}

		var ok bool
		if value, ok = fValue.([]string); !ok {
			var item string
			if item, ok = fValue.(string); !ok {
				return value, ExpressionValueInvalidError{
					Declaration: expr.GetDeclaration(), Type: expr.GetType()}
			}
			return []string{item}, nil
		}
		return value, nil
	}

	// expr with serveral sub expr
	for _, subExpr := range expr.Literal {
		subValue, err := subExpr.GetValue(stack)
		if err != nil {
			return nil, errors.Trace(err)
		}
		value = append(value, subValue)
	}
	return value, nil
}

// Select returns value of list expression by index
func (expr *StringListExpr) Select(stack utils.StackInterface, index int) (interface{}, error) {
	value, err := expr.GetValue(stack)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if index >= len(value) || index < 0 {
		return nil, errors.Errorf("invalid index %d for value of %s", index, expr.GetDeclaration())
	}
	return value[index], nil
}

// UnmarshalJSON sets the object from the provided JSON representation
func (expr *StringListExpr) UnmarshalJSON(data []byte) error {
	var v []*StringExpr
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

	// Perhaps we have a single item, like "foo" which
	// occurs occasionally.
	var v2 StringExpr
	err3 := json.Unmarshal(data, &v2)
	if err3 == nil {
		expr.Func = nil
		expr.Literal = []*StringExpr{&v2}
		return nil
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}
