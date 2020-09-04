package openapiclient

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type parseOpenAPISpecSuite struct {
	suite.Suite
}

func (s *parseOpenAPISpecSuite) TestParseOpenAPISpec() {
	c := new(client)
	spec := `{
    "openapi": "3.0.0",
    "info": {
        "version": "SDS_4.2.009.0",
        "contact": {},
        "description": "XMS is the controller of distributed storage system",
        "license": {
            "name": "Commercial"
        },
        "title": "XMS API"
    },
    "paths": {
        "/os-replication-paths/": {
            "get": {
                "parameters": [
                    {
                        "description": "paging param",
                        "name": "limit",
                        "in": "query",
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                        }
                    },
                    {
                        "description": "paging param",
                        "name": "offset",
                        "in": "query",
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                    	}
					}
				],
				"operationId": "op1"
			},
			"put": {
                "parameters": [
                    {
                        "description": "paging param",
                        "name": "test",
                        "in": "path",
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                        }
                    },
                    {
                        "description": "paging param",
                        "name": "offset",
                        "in": "query",
                        "schema": {
                            "type": "integer",
                            "format": "int64"
                    	}
					}
				],
				"operationId": "op2"
			}
		},
		"/osss-test/": {
			"post": {
				"parameters": [
					{
						"name": "id",
						"in": "path"
					}
				],
				"operationId": "test-osss"
			}
		}
	}
}`
	assert.NoError(s.T(), c.ParseOpenAPISpec([]byte(spec)))
	expOpenAPIInfo := &openAPIInfo{
		OpenAPI: "3.0.0",
		Version: "SDS_4.2.009.0",
		Paths: &openAPI{
			OperationIDs: map[string]*openAPIMethodInfo{
				"op1": {
					URL:         "/os-replication-paths/",
					Method:      "GET",
					OperationID: "op1",
				},
				"op2": {
					URL:         "/os-replication-paths/",
					Method:      "PUT",
					PathParams:  []string{"test"},
					OperationID: "op2",
				},
				"test-osss": {
					URL:         "/osss-test/",
					Method:      "POST",
					PathParams:  []string{"id"},
					OperationID: "test-osss",
				},
			},
		},
	}
	assert.Equal(s.T(), expOpenAPIInfo, c.openAPI)
}

type stringReadCloser struct {
	*strings.Reader
}

func (*stringReadCloser) Close() error {
	return nil
}

func TestParseOpenAPISpec(t *testing.T) {
	suite.Run(t, new(parseOpenAPISpecSuite))
}

type mockedHTTPClient struct {
	mock.Mock

	respData string
}

func (c *mockedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (c *mockedHTTPClient) Get(url string) (*http.Response, error) {
	args := c.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

type callAPISuite struct {
	suite.Suite

	apiClient *client
}

func (s *callAPISuite) SetupTest() {
	apiInfo := &openAPIInfo{
		OpenAPI: "3.0.0",
		Version: "SDS_4.2.009.0",
		Paths: &openAPI{
			OperationIDs: map[string]*openAPIMethodInfo{
				"op1": {
					URL:         "/os-replication-paths/",
					Method:      "GET",
					OperationID: "op1",
				},
				"op2": {
					URL:         "/os-replication-paths/",
					Method:      "PUT",
					PathParams:  []string{"test"},
					OperationID: "op2",
				},
				"test-osss": {
					URL:         "/osss-test/",
					Method:      "POST",
					PathParams:  []string{"id"},
					OperationID: "test-osss",
				},
			},
		},
	}
	s.apiClient = new(client)
	s.apiClient.SetServer("http://1.1.1.1")
	s.apiClient.openAPI = apiInfo
}

func (s *callAPISuite) TestCallAPI() {
	mockedClient := new(mockedHTTPClient)
	s.apiClient.httpClient = mockedClient
	mockedClient.On("Do", mock.AnythingOfType("*http.Request")).
		Return(&http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("{}")),
		}, nil)

	_, err := s.apiClient.CallAPI("op1", nil, nil)
	assert.NoError(s.T(), err)

	mockedClient.AssertCalled(s.T(), "Do", mock.AnythingOfType("*http.Request"))
}

func (s *callAPISuite) TestCallWithouPathParam() {
	_, err := s.apiClient.CallAPI("op2", nil, nil)
	assert.Error(s.T(), err)

	assert.EqualError(s.T(), err, "path param test not set")
}

func (s *callAPISuite) TestWithNotFoundOperationID() {
	_, err := s.apiClient.CallAPI("test33", nil, nil)
	assert.EqualError(s.T(), err, "operation id test33 not found")
}

func TestCallAPI(t *testing.T) {
	suite.Run(t, new(callAPISuite))
}
