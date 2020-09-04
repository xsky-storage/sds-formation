package parser

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// IntegerListExpr is a integer list expression
type IntegerListExpr struct {
	baseExpr
	Literal []*IntegerExpr
}

// GetType returns type of integer list expression
func (expr *IntegerListExpr) GetType() string {
	return ValueTypeIntegerList
}

// IsReady returns if integer list expression is ready
func (expr *IntegerListExpr) IsReady(stack utils.StackInterface) (ready bool) {
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

// GetValue returns value of integer list expression
func (expr *IntegerListExpr) GetValue(stack utils.StackInterface) (value []int64, err error) {
	if expr.Func != nil {
		fValue, err := expr.Func.getValue(stack)
		if err != nil {
			return value, errors.Trace(err)
		}

		var ok bool
		if value, ok = fValue.([]int64); !ok {
			var item int64
			if item, ok = fValue.(int64); !ok {
				return value, ExpressionValueInvalidError{
					Declaration: expr.GetDeclaration(), Type: expr.GetType()}
			}
			return []int64{item}, nil
		}
		return value, nil
	}

	// expr with serveral sub expr
	for _, subExpr := range expr.Literal {
		subValue, e := subExpr.GetValue(stack)
		if e != nil {
			return nil, errors.Trace(e)
		}
		value = append(value, subValue)
	}
	return value, nil
}

// Select returns value of list expression by index
func (expr *IntegerListExpr) Select(stack utils.StackInterface, index int) (interface{}, error) {
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
func (expr *IntegerListExpr) UnmarshalJSON(data []byte) error {
	var v []*IntegerExpr
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
	var v2 IntegerExpr
	err3 := json.Unmarshal(data, &v2)
	if err3 == nil {
		expr.Func = nil
		expr.Literal = []*IntegerExpr{&v2}
		return nil
	}

	// Return the original error trying to unmarshal the literal expression,
	// which will be the most expressive.
	return err
}
