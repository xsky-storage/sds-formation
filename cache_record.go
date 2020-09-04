package formation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

// CacheRecord defines struct of resource create cache record
type CacheRecord struct {
	Name         string
	ResourceType string
	InTemplate   bool
	ValueType    string
	Value        json.RawMessage
}

func (r *CacheRecord) String() string {
	return fmt.Sprintf("CacheRecord(ResourceName:%s, ResourceType:%s)", r.Name, r.ResourceType)
}

func (r *CacheRecord) getValueType(typ string) reflect.Type {
	switch {
	case typ == "string":
		return reflect.TypeOf("")
	case strings.HasPrefix(typ, "int"):
		return reflect.TypeOf(int64(0))
	case strings.HasPrefix(typ, "float"):
		return reflect.TypeOf(float64(0))
	case typ == "bool":
		return reflect.TypeOf(false)
	}
	return nil
}

func (r *CacheRecord) getContainerType(typ string) reflect.Type {
	switch {
	case strings.HasPrefix(typ, "[]"):
		elemTyp := r.getValueType(strings.TrimPrefix(typ, "[]"))
		if elemTyp == nil {
			return nil
		}
		return reflect.SliceOf(elemTyp)
	}
	return r.getValueType(typ)
}

// GetExpr returns resource's expr in cache record
func (r *CacheRecord) GetExpr() (interface{}, error) {
	if r.ResourceType == utils.ResourceTemplate {
		val, err := r.GetTemplateExpr()
		if err != nil {
			return nil, errors.Trace(err)
		}
		return val, nil
	}
	val, err := r.GetExprWithTypeValue(r.Value, r.ValueType)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return val, nil
}

// GetExprWithTypeValue returns resource's expr in cache record with specifics type and value
func (r *CacheRecord) GetExprWithTypeValue(value json.RawMessage, typeExpr string) (interface{}, error) {
	valTyp := r.getContainerType(typeExpr)
	if valTyp == nil {
		return nil, errors.Errorf("invalid value type found in cache record: %s", typeExpr)
	}
	val := reflect.New(valTyp).Interface()
	if err := json.Unmarshal(value, val); err != nil {
		return nil, errors.Trace(err)
	}
	return reflect.Indirect(reflect.ValueOf(val)).Interface(), nil
}

func (r *CacheRecord) getTemplateValueType(typ string) ([]map[string]string, error) {
	rawTypeBytes := json.RawMessage(typ)
	var types []map[string]string
	if err := json.Unmarshal(rawTypeBytes, &types); err != nil {
		return nil, errors.Trace(err)
	}
	return types, nil
}

// GetTemplateExpr load repr from cache record for template resource
func (r *CacheRecord) GetTemplateExpr() ([]map[string]interface{}, error) {
	types, err := r.getTemplateValueType(r.ValueType)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rawData := []map[string]json.RawMessage{}
	if err = json.Unmarshal(r.Value, &rawData); err != nil {
		return nil, errors.Trace(err)
	}
	if len(types) != len(rawData) {
		return nil, errors.Errorf("got invalid cache data of template resource %s", r.Name)
	}
	realVals := []map[string]interface{}{}
	for i, datas := range rawData {
		typesI := types[i]
		if len(datas) != len(typesI) {
			return nil, errors.Errorf("got invalid cache data of template resource %s", r.Name)
		}
		valsMap := map[string]interface{}{}
		for name, valData := range datas {
			typeExpr, ok := typesI[name]
			if !ok {
				return nil, errors.Errorf("got invalid cache data of template resource %s", r.Name)
			}
			val, err := r.GetExprWithTypeValue(valData, typeExpr)
			if err != nil {
				return nil, errors.Trace(err)
			}
			valsMap[name] = val
		}
		realVals = append(realVals, valsMap)
	}
	return realVals, nil
}

func encodeTemplateCacheType(value interface{}, resourceName string) (string, error) {
	values, ok := value.([]map[string]interface{})
	if !ok {
		return "", errors.Errorf("got invalid value of template resource %s", resourceName)
	}
	valTypes := []map[string]string{}
	for _, value := range values {
		typeMap := map[string]string{}
		for name, val := range value {
			typeMap[name] = reflect.TypeOf(val).String()
		}
		valTypes = append(valTypes, typeMap)
	}
	typeData, err := json.Marshal(valTypes)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(typeData), nil
}

// GetCacheRecord returns a cache record object
func GetCacheRecord(resourceName, resourceType string, value interface{}, inTmpl bool) (
	*CacheRecord, error) {

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, errors.Trace(err)
	}
	valueType := reflect.TypeOf(value).String()
	if resourceType == utils.ResourceTemplate {
		valueType, err = encodeTemplateCacheType(value, resourceName)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	return &CacheRecord{
		Name:         resourceName,
		ResourceType: resourceType,
		ValueType:    valueType,
		InTemplate:   inTmpl,
		Value:        json.RawMessage(valueBytes),
	}, nil
}
