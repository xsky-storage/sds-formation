package formation

import (
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// IntegerList resource
type IntegerList struct {
	ResourceBase
	Attributes []*parser.IntegerListExpr
}

// Init inits resource instance
func (integerList *IntegerList) Init(stack utils.StackInterface) {
	integerList.ResourceBase.Init(stack)
	integerList.setDelegate(integerList)
}

// GetType return resource type
func (integerList *IntegerList) GetType() string {
	return utils.ResourceIntegerList
}

// IsReady check if the formation args are ready
func (integerList *IntegerList) IsReady() (ready bool) {
	for _, attr := range integerList.Attributes {
		if !integerList.isReady(attr) {
			return false
		}
	}

	return true
}

// Create create the resource
func (integerList *IntegerList) Create() (created bool, err error) {
	repr := []int64{}
	for _, attr := range integerList.Attributes {
		val := integerList.getIntegerListValue(attr)
		repr = append(repr, val...)
	}
	integerList.repr = repr
	created = true

	return
}
