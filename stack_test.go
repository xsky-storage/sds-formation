package formation

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"xsky.com/sds-formation/tests"
)

type stackLoadCacheSuite struct {
	suite.Suite

	stack       *Stack
	mockedFile  *tests.MockedFile
	oldOpenFile OpenFileFunc
}

func (s *stackLoadCacheSuite) SetupTest() {
	s.oldOpenFile = OpenFile
	s.mockedFile = new(tests.MockedFile)
	OpenFile = func(string, int, ...os.FileMode) (io.ReadWriteCloser, error) {
		return s.mockedFile, nil
	}
	s.stack = new(Stack)
}

func (s *stackLoadCacheSuite) TeardownTest() {
	OpenFile = s.oldOpenFile
}

func (s *stackLoadCacheSuite) TestWithNoCacheFile() {
	s.mockedFile.On("Read", mock.AnythingOfType("[]uint8"))

	assert.NoError(s.T(), s.stack.loadCache("test"))

	s.mockedFile.AssertCalled(s.T(), "Read", mock.AnythingOfType("[]uint8"))
}

func (s *stackLoadCacheSuite) TestLoadCache() {
	cacheData := `{"Name":"token","ResourceType":"Token","ValueType":"string","Value":"17412dde75c34e92ad7d931bb4b2c287"}
{"Name":"disk_list","ResourceType":"DiskList","ValueType":"[]int64","Value":[1,2]}
`
	s.mockedFile.SetReadData([]byte(cacheData))
	s.mockedFile.On("Read", mock.AnythingOfType("[]uint8"))

	assert.NoError(s.T(), s.stack.loadCache("test"))

	s.mockedFile.AssertCalled(s.T(), "Read", mock.AnythingOfType("[]uint8"))
	expCacheRecords := []*CacheRecord{
		{
			Name:         "token",
			ResourceType: "Token",
			ValueType:    "string",
			Value:        json.RawMessage("\"17412dde75c34e92ad7d931bb4b2c287\""),
		},
		{
			Name:         "disk_list",
			ResourceType: "DiskList",
			ValueType:    "[]int64",
			Value:        json.RawMessage("[1,2]"),
		},
	}
	assert.Equal(s.T(), expCacheRecords, s.stack.cacheExprs)
	assert.NotNil(s.T(), s.stack.cacheFile)
}

func TestStackLoadCacheSuite(t *testing.T) {
	suite.Run(t, new(stackLoadCacheSuite))
}

type makeTemplateResourceContextSuite struct {
	suite.Suite

	stack *Stack
}

func (s *makeTemplateResourceContextSuite) SetupTest() {
	s.stack = new(Stack)
}

func (s *makeTemplateResourceContextSuite) TestMakeTemplateContext() {
	templatesStr := `
	[
		{
			"Name":         "tmpl",
			"Type":         "Template",
			"TemplateName": "xx"
		},
		{
			"Name":           "tmpl1",
			"Type":           "Template",
			"TemplateName":   "xx",
			"Context": [
				{
					"Name":  "name",
					"Type":  "Integer",
					"Value": 12
				},
				{
					"Name":  "type",
					"Type":  "String",
					"Value": "test"
				}
			]
		},
		{
			"Name":         "tmpl2",
			"Type":         "Template",
			"TemplateName": "xx",
			"Context": [
				{
					"Name":  "name",
					"Type":  "IntegerList",
					"Value": [12, 1, 2],
					"Action": "range"
				},
				{
					"Name":  "type",
					"Type":  "String",
					"Value": "type"
				}
			]
		},
		{
			"Name":         "tmpl3",
			"Type":         "Template",
			"TemplateName": "xx",
			"Context": [
				{
					"Name":  "name",
					"Type":  "IntegerList",
					"Value": [12, 99],
					"Action": "range"
				},
				{
					"Name":  "type",
					"Type":  "StringList",
					"Value": ["type1", "type2"],
					"Action": "range"
				},
				{
					"Name":  "role",
					"Type":  "String",
					"Value": "role1"
				}
			]
		}
	]
	`
	testTemplates := make([]*ResourceInTemplate, 0)
	s.NoError(json.Unmarshal([]byte(templatesStr), &testTemplates))
	contextsExp := [][]map[string]interface{}{
		{
			{},
		},
		{
			{
				"name": int64(12),
				"type": "test",
			},
		},
		{
			{
				"name": int64(12),
				"type": "type",
			},
			{
				"name": int64(1),
				"type": "type",
			},
			{
				"name": int64(2),
				"type": "type",
			},
		},
		{
			{
				"name": int64(12),
				"type": "type1",
				"role": "role1",
			},
			{
				"name": int64(12),
				"type": "type2",
				"role": "role1",
			},
			{
				"name": int64(99),
				"type": "type1",
				"role": "role1",
			},
			{
				"name": int64(99),
				"type": "type2",
				"role": "role1",
			},
		},
	}

	for i, template := range testTemplates {
		contextList, err := s.stack.CheckGetTemplateResourceContext(template)
		s.NoError(err)

		s.Equal(contextsExp[i], contextList)
	}
}

func (s *makeTemplateResourceContextSuite) TestWithFuncInContextVal() {
	s.stack.resourceValueMap = map[string]interface{}{
		"context1_val": int64(12),
		"context2_val": []string{"val1", "val2", "val3"},
		"context3_val": []int64{5, 6},
	}
	templateStr := `{
		"Name": "tmpl",
		"Type": "Template",
		"TemplateName": "tmpl1",
		"Context": [
			{
				"Name": "context1",
				"Type": "Integer",
				"Value": {"Ref": "context1_val"}
			},
			{
				"Name": "context2",
				"Type": "String",
				"Value": {"Select": [1, {"Ref": "context2_val"}]}
			},
			{
				"Name": "context3",
				"Type": "IntegerList",
				"Value": {"Ref": "context3_val"},
				"Action": "range"
			}
		]
	}`
	template := new(ResourceInTemplate)
	s.NoError(json.Unmarshal([]byte(templateStr), template))

	contextList, err := s.stack.CheckGetTemplateResourceContext(template)
	s.NoError(err)

	contextListExpr := []map[string]interface{}{
		{
			"context1": int64(12),
			"context2": "val2",
			"context3": int64(5),
		},
		{
			"context1": int64(12),
			"context2": "val2",
			"context3": int64(6),
		},
	}
	s.Equal(contextListExpr, contextList)
}

func (s *makeTemplateResourceContextSuite) TestWithError() {
	templateStr := `{
		"Name": "temp",
		"Type": "Template",
		"TemplateName": "tmpl1",
		"Context": [
			{
				"Name": "context1",
				"Type": "Integer",
				"Value": {"Ref": "context"}
			}
		]
	}`
	template := new(ResourceInTemplate)
	s.NoError(json.Unmarshal([]byte(templateStr), template))

	// test context not ready
	_, err := s.stack.CheckGetTemplateResourceContext(template)
	s.EqualError(err, "template resource temp's context context1 not found")
}

func TestMakeTemplateResourceContextSuite(t *testing.T) {
	suite.Run(t, new(makeTemplateResourceContextSuite))
}
