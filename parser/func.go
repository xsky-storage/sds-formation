package parser

import (
	"encoding/json"
	"reflect"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// consts for functoin names
const (
	FuncNameRef              = "Ref"
	FuncNameSelect           = "Select"
	FuncNameTemplateAttrElem = "TemplateAttrElem"
	FuncNameTemplateAttr     = "TemplateAttr"
)

// Func defines function interface
type Func interface {
	isReady(stack utils.StackInterface) (ready bool)
	getValue(stack utils.StackInterface) (value interface{}, err error)
}

// RefFunc defines function that returns returns the value of the specified parameter or resource
type RefFunc struct {
	Ref string `json:"Ref"`
}

func (refFunc *RefFunc) isReady(stack utils.StackInterface) (ready bool) {
	return stack.GetResourceValue(refFunc.Ref) != nil
}

func (refFunc *RefFunc) getValue(stack utils.StackInterface) (value interface{}, err error) {
	value = stack.GetResourceValue(refFunc.Ref)
	if value == nil {
		return nil, ExpressionParamInsufficientError{ParamName: refFunc.Ref}
	}
	return value, nil
}

func (refFunc *RefFunc) unmarshal(valueType string, data json.RawMessage) (err error) {
	if err = json.Unmarshal(data, &refFunc.Ref); err != nil {
		return errors.Trace(err)
	}
	return
}

// SelectFunc defines function that returns a single object from a list of objects by index
type SelectFunc struct {
	Index    int
	ListExpr ListExprType
}

func (selectFunc *SelectFunc) isReady(stack utils.StackInterface) (ready bool) {
	return selectFunc.ListExpr.IsReady(stack)
}

func (selectFunc *SelectFunc) getValue(stack utils.StackInterface) (
	value interface{}, err error) {

	if value, err = selectFunc.ListExpr.Select(stack, selectFunc.Index); err != nil {
		return nil, errors.Trace(err)
	}
	return
}

func (selectFunc *SelectFunc) unmarshal(valueType string, data json.RawMessage) (err error) {
	rawMessages := []json.RawMessage{}
	if err = json.Unmarshal(data, &rawMessages); err != nil {
		return errors.Trace(err)
	}
	if len(rawMessages) != 2 {
		return errors.Errorf("cannot decode fuction")
	}
	if err = json.Unmarshal(rawMessages[0], &selectFunc.Index); err != nil {
		return errors.Errorf("cannot decode function")
	}

	switch valueType {
	case ValueTypeInteger:
		selectFunc.ListExpr = new(IntegerListExpr)
		if err := json.Unmarshal(rawMessages[1], selectFunc.ListExpr); err != nil {
			return errors.Trace(err)
		}
	case ValueTypeString:
		selectFunc.ListExpr = new(StringListExpr)
		if err := json.Unmarshal(rawMessages[1], selectFunc.ListExpr); err != nil {
			return errors.Trace(err)
		}
	default:
		return errors.Errorf("cannot decode function")
	}
	return
}

// TemplateAttrElemenFunc defines function that returns a template resource's attr value
type TemplateAttrElemenFunc struct {
	Ref       string
	Attr      string
	Index     int
	ValueType string
}

func (templateAttrElemenFunc *TemplateAttrElemenFunc) isReady(stack utils.StackInterface) (ready bool) {
	tmplReprData := stack.GetResourceValue(templateAttrElemenFunc.Ref)
	if tmplReprData == nil {
		return false
	}
	tmplRepr, ok := tmplReprData.([]map[string]interface{})
	if !ok {
		return false
	}
	if len(tmplRepr) <= templateAttrElemenFunc.Index {
		return false
	}
	_, ready = tmplRepr[templateAttrElemenFunc.Index][templateAttrElemenFunc.Attr]
	return
}

func (templateAttrElemenFunc *TemplateAttrElemenFunc) getValue(stack utils.StackInterface) (
	value interface{}, err error) {

	tmplReprData := stack.GetResourceValue(templateAttrElemenFunc.Ref)
	if tmplReprData == nil {
		return nil, errors.NotFoundf("template %s's repr", templateAttrElemenFunc.Ref)
	}
	tmplRepr, ok := tmplReprData.([]map[string]interface{})
	if !ok {
		return nil, errors.Errorf("invalid repr type of %s", templateAttrElemenFunc.Ref)
	}
	return tmplRepr[templateAttrElemenFunc.Index][templateAttrElemenFunc.Attr], nil
}

func (templateAttrElemenFunc *TemplateAttrElemenFunc) unmarshal(valueType string, data json.RawMessage) (err error) {
	if err = json.Unmarshal(data, &templateAttrElemenFunc); err != nil {
		return errors.Trace(err)
	}
	templateAttrElemenFunc.ValueType = valueType
	return nil
}

// TemplateAttrFunc defines function that returns a template resource's attr value
type TemplateAttrFunc struct {
	Ref       string
	Attr      string
	ValueType string
}

func (templateAttrFunc *TemplateAttrFunc) isReady(stack utils.StackInterface) (ready bool) {
	tmplReprData := stack.GetResourceValue(templateAttrFunc.Ref)
	if tmplReprData == nil {
		return false
	}
	tmplRepr, ok := tmplReprData.([]map[string]interface{})
	if !ok {
		return false
	}
	for _, reprs := range tmplRepr {
		if reprs[templateAttrFunc.Attr] != nil {
			return true
		}
	}
	return false
}

func (templateAttrFunc *TemplateAttrFunc) getValue(stack utils.StackInterface) (
	value interface{}, err error) {

	tmplReprData := stack.GetResourceValue(templateAttrFunc.Ref)
	if tmplReprData == nil {
		return nil, errors.NotFoundf("template %s's repr", templateAttrFunc.Ref)
	}
	tmplRepr, ok := tmplReprData.([]map[string]interface{})
	if !ok {
		return nil, errors.Errorf("invalid repr type of %s", templateAttrFunc.Ref)
	}
	var values reflect.Value
	switch templateAttrFunc.ValueType {
	case "StringList":
		values = reflect.ValueOf([]string{})
	case "IntegerList":
		values = reflect.ValueOf([]int64{})
	}
	for _, reprs := range tmplRepr {
		repr := reprs[templateAttrFunc.Attr]
		if repr == nil {
			continue
		}
		val := reflect.ValueOf(repr)
		if val.Kind() == reflect.Slice {
			for i := 0; i < val.Len(); i++ {
				values = reflect.Append(values, val.Index(i))
			}
		} else {
			values = reflect.Append(values, val)
		}
	}
	return values.Interface(), nil
}

func (templateAttrFunc *TemplateAttrFunc) unmarshal(valueType string, data json.RawMessage) (err error) {
	if err = json.Unmarshal(data, &templateAttrFunc); err != nil {
		return errors.Trace(err)
	}
	templateAttrFunc.ValueType = valueType
	return nil
}

func unmarshalFunc(valueType string, data []byte) (Func, error) {
	rawDecode := map[string]json.RawMessage{}
	err := json.Unmarshal(data, &rawDecode)
	if err != nil {
		return nil, err
	}
	for funcName, funcData := range rawDecode {
		switch funcName {
		case FuncNameRef:
			f := new(RefFunc)
			if err = f.unmarshal(valueType, funcData); err == nil {
				return f, nil
			}
		case FuncNameSelect:
			f := new(SelectFunc)
			if err = f.unmarshal(valueType, funcData); err == nil {
				return f, nil
			}
		case FuncNameTemplateAttrElem:
			f := new(TemplateAttrElemenFunc)
			if err = f.unmarshal(valueType, funcData); err == nil {
				return f, nil
			}
		case FuncNameTemplateAttr:
			f := new(TemplateAttrFunc)
			if valueType != ValueTypeStringList && valueType != ValueTypeIntegerList {
				errors.Errorf("TemplateAttr func's value type can't be %s", valueType)
			}
			if err = f.unmarshal(valueType, funcData); err == nil {
				return f, nil
			}
		default:
			return nil, FunctionUnknownError{FunctionName: funcName}
		}
	}
	return nil, errors.Errorf("cannot decode function")
}
