package formation

import (
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// StringList resource
type StringList struct {
	ResourceBase
	Attributes []*parser.StringListExpr
}

// Init inits reosource instance
func (stringList *StringList) Init(stack utils.StackInterface) {
	stringList.ResourceBase.Init(stack)
	stringList.setDelegate(stringList)
}

// GetType return resource type
func (stringList *StringList) GetType() string {
	return utils.ResourceStringList
}

// IsReady check if the formation args are ready
func (stringList *StringList) IsReady() (ready bool) {
	for _, attr := range stringList.Attributes {
		if !stringList.isReady(attr) {
			return false
		}
	}

	return true
}

// Create create the resource
func (stringList *StringList) Create() (created bool, err error) {
	repr := []string{}
	for _, attr := range stringList.Attributes {
		val := stringList.getStringListValue(attr)
		repr = append(repr, val...)
	}

	stringList.repr = repr
	return true, nil
}
