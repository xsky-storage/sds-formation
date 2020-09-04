package formation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"xsky.com/sds-formation/utils"
)

type cacheRecordSuite struct {
	suite.Suite
}

func (s *cacheRecordSuite) TestGetRecordExpr() {
	testFunc := func(valType, value string, expVal interface{}) {
		record := &CacheRecord{
			Name:         "test",
			ResourceType: "test",
			ValueType:    valType,
			Value:        json.RawMessage(value),
		}

		val, err := record.GetExpr()
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), expVal, val)
	}

	testFunc("int", "1", int64(1))
	testFunc("[]int", "[1]", []int64{1})
	testFunc("float64", "1", float64(1))
	testFunc("[]float64", "[1]", []float64{1})
	testFunc("string", "\"test\"", "test")
	testFunc("[]string", "[\"test1\", \"test2\"]", []string{"test1", "test2"})
}

func (s *cacheRecordSuite) TestGetExprWithUnsupportedValueType() {
	record := &CacheRecord{
		Name:         "test",
		ResourceType: "test",
		ValueType:    "map[int]int",
		Value:        json.RawMessage("{\"a\": \"b\"}"),
	}
	_, err := record.GetExpr()
	assert.Error(s.T(), err)
}

func (s *cacheRecordSuite) TestGetTemplateResourceCache() {
	val := []map[string]interface{}{
		{
			"a": int64(12),
			"b": []int64{1, 2, 3, 4},
			"c": false,
		},
		{
			"d": "qwer",
			"e": []string{"qwe", "zxc", "ad"},
		},
	}
	cacheRecord, err := GetCacheRecord("test", utils.ResourceTemplate, val, true)
	s.NoError(err)

	valActual, err := cacheRecord.GetTemplateExpr()
	s.NoError(err)
	s.Equal(val, valActual)
}

func TestCacheRecordSuite(t *testing.T) {
	suite.Run(t, new(cacheRecordSuite))
}
