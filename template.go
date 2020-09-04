package formation

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/parser"
	resources "xsky.com/sds-formation/resources"
	"xsky.com/sds-formation/utils"
)

var inTemplate = false

// Template resource template
type Template struct {
	Description string                     `json:",omitempty"`
	Parameters  map[string]*Parameter      `json:",omitempty"`
	Resources   []*ResourceInTemplate      `json:",omitempty"`
	Templates   map[string]json.RawMessage `json:",omitempty"`
}

// CheckTemplates check resources templates is valid
func (t *Template) CheckTemplates() error {
	// nested tempalte not support currently, this check may removed in future
	inTemplate = true
	defer func() {
		inTemplate = false
	}()
	for templateName, templateData := range t.Templates {
		tmpResurces := make([]*ResourceInTemplate, 0)
		if err := json.Unmarshal(templateData, &tmpResurces); err != nil {
			return errors.Annotatef(err, "in template %s", templateName)
		}
	}
	return nil
}

// ResourceInTemplate a resource in a template
type ResourceInTemplate struct {
	Name          string
	Type          string
	Action        string
	Sleep         int
	WaitInterval  int
	CheckInterval int
	Context       []*templateContext
	TemplateName  string
	Properties    utils.ResourceInterface
}

type templateContext struct {
	Name   string          `json:",omitempty"`
	Type   string          `json:",omitempty"` // String,StringList,Integer,IntegerList
	Action string          `json:",omitempty"` // range
	Value  parser.ExprType `json:",omitempty"`
}

// UnmarshalJSON sets the template context from the provided JSON
func (t *templateContext) UnmarshalJSON(buf []byte) (err error) {
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return errors.Trace(err)
	}

	nameBytes, ok := m["Name"]
	if !ok {
		return errors.Errorf("Name is required for template context")
	}
	if err = json.Unmarshal(nameBytes, &t.Name); err != nil {
		return errors.Trace(err)
	}

	typeBytes, ok := m["Type"]
	if !ok {
		return errors.Errorf("Type is required for template context")
	}
	if err = json.Unmarshal(typeBytes, &t.Type); err != nil {
		return errors.Trace(err)
	}

	actionBytes, ok := m["Action"]
	if ok {
		if err = json.Unmarshal(actionBytes, &t.Action); err != nil {
			return errors.Trace(err)
		}
	}
	if t.Action != "" && t.Action != utils.ContextValueActionRange {
		return errors.Errorf("got invalid template context action %s", t.Action)
	}

	valueBytes, ok := m["Value"]
	if !ok {
		return errors.Errorf("Value is required for template context")
	}
	switch t.Type {
	case utils.ContextTypeInteger:
		t.Value = new(parser.IntegerExpr)
	case utils.ContextTypeIntegerList:
		t.Value = new(parser.IntegerListExpr)
	case utils.ContextTypeStringList:
		t.Value = new(parser.StringListExpr)
	case utils.ContextTypeString:
		t.Value = new(parser.StringExpr)
	case utils.ContextTypeBool:
		t.Value = new(parser.BoolExpr)
	default:
		return errors.Errorf("got invalid template context type %s", t.Type)
	}
	if err = json.Unmarshal(valueBytes, t.Value); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// UnmarshalJSON sets the object from the provided JSON representation
func (r *ResourceInTemplate) UnmarshalJSON(buf []byte) (err error) {
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return errors.Trace(err)
	}

	nameBytes, ok := m["Name"]
	if !ok {
		return errors.Errorf("Name is required for a resource")
	}
	if err = json.Unmarshal(nameBytes, &r.Name); err != nil {
		return errors.Trace(err)
	}

	typeBytes, ok := m["Type"]
	if !ok {
		return errors.Errorf("Type is required for a resource")
	}
	if err = json.Unmarshal(typeBytes, &r.Type); err != nil {
		return errors.Trace(err)
	}

	if inTemplate && r.Type == utils.ResourceTemplate {
		return errors.New("nested template unsupported")
	}

	templateNameBytes, ok := m["TemplateName"]
	if !ok && r.Type == utils.ResourceTemplate {
		return errors.Errorf("TemplateName is required for tempalte resource")
	} else if ok {
		if err = json.Unmarshal(templateNameBytes, &r.TemplateName); err != nil {
			return errors.Trace(err)
		}
	}

	contextBytes, ok := m["Context"]
	if ok {
		if err = json.Unmarshal(contextBytes, &r.Context); err != nil {
			return errors.Trace(err)
		}
	}

	actionBytes, ok := m["Action"]
	if ok {
		if err = json.Unmarshal(actionBytes, &r.Action); err != nil {
			return errors.Trace(err)
		}
	}

	sleepBytes, ok := m["Sleep"]
	if ok {
		if err = json.Unmarshal(sleepBytes, &r.Sleep); err != nil {
			return errors.Trace(err)
		}
	}

	waitIntervalBytes, ok := m["WaitInterval"]
	if ok {
		if err = json.Unmarshal(waitIntervalBytes, &r.WaitInterval); err != nil {
			return errors.Trace(err)
		}
	}

	checkIntervalBytes, ok := m["CheckInterval"]
	if ok {
		if err = json.Unmarshal(checkIntervalBytes, &r.CheckInterval); err != nil {
			return errors.Trace(err)
		}
	}

	r.Properties = resources.NewResource(r.Type, r.Action)
	if r.Properties == nil {
		return errors.Errorf("unknown resource type: %s", r.Type)
	}
	properties, ok := m["Properties"]
	if r.Type != utils.ResourceTemplate {
		if !ok {
			return errors.Errorf("Properties is requried for a resource")
		}
		if err := json.Unmarshal(properties, r.Properties); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// Parameter parameter in template
type Parameter struct {
	Type  string      `json:",omitempty"`
	Value interface{} `json:",omitempty"`
}

// UnmarshalJSON interface
func (p *Parameter) UnmarshalJSON(buf []byte) (err error) {
	m := map[string]interface{}{}
	if err = json.Unmarshal(buf, &m); err != nil {
		return err
	}
	defaultBuf, _ := json.Marshal(m["Value"])
	if err != nil {
		return err
	}

	p.Type = m["Type"].(string)
	switch p.Type {
	case "Integer":
		var integer int64
		err = json.Unmarshal(defaultBuf, &integer)
		p.Value = integer
	case "String":
		var str string
		err = json.Unmarshal(defaultBuf, &str)
		p.Value = str
	case "IntegerList":
		integerList := []int64{}
		err = json.Unmarshal(defaultBuf, &integerList)
		p.Value = integerList
	case "StringList":
		strList := []string{}
		err = json.Unmarshal(defaultBuf, &strList)
		p.Value = strList

	default:
		err = errors.Errorf("unknown parameter type %s", p.Type)
	}
	if err != nil {
		return errors.Trace(err)
	}

	return
}
