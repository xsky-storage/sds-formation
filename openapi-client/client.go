package openapiclient

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/juju/errors"
)

type openAPIMethodInfo struct {
	URL         string
	Method      string
	PathParams  []string
	OperationID string
}

type openAPI struct {
	OperationIDs map[string]*openAPIMethodInfo
}

// UnmarshalJSON implements json Unmarshaler
func (o *openAPI) UnmarshalJSON(bytes []byte) error {
	o.OperationIDs = make(map[string]*openAPIMethodInfo)
	pathMap := map[string]json.RawMessage{}
	err := json.Unmarshal(bytes, &pathMap)
	if err != nil {
		return errors.Trace(err)
	}
	for path, rawPathInfo := range pathMap {
		pathInfo := map[string]*struct {
			OperationID string `json:"operationId"`
			Params      []*struct {
				Name string `json:"name"`
				In   string `json:"in"`
			} `json:"parameters"`
		}{}
		if err = json.Unmarshal(rawPathInfo, &pathInfo); err != nil {
			return errors.Trace(err)
		}
		for method, methodInfo := range pathInfo {
			info := &openAPIMethodInfo{
				URL:         path,
				Method:      strings.ToUpper(method),
				OperationID: methodInfo.OperationID,
			}
			for _, param := range methodInfo.Params {
				if param.In == "path" {
					info.PathParams = append(info.PathParams, param.Name)
				}
			}
			o.OperationIDs[info.OperationID] = info
		}
	}
	return nil
}

type openAPIInfo struct {
	OpenAPI string   `json:"openapi"`
	Version string   `json:"version"`
	Paths   *openAPI `json:"paths"`
}

// UnmarshalJSON implements json Unmarshaller
func (o *openAPIInfo) UnmarshalJSON(bytes []byte) error {
	apiInfo := new(struct {
		OpenAPI string `json:"openapi"`
		Info    struct {
			Version string `json:"version"`
		} `json:"info"`
		Paths *openAPI `json:"paths"`
	})
	err := json.Unmarshal(bytes, &apiInfo)
	if err != nil {
		return errors.Trace(err)
	}
	o.OpenAPI = apiInfo.OpenAPI
	o.Version = apiInfo.Info.Version
	o.Paths = apiInfo.Paths

	return nil
}

// Client defines interface of openapi client
type Client interface {
	Init() error
	SetServer(string)
	SetToken(string)
	LoadSpec() error
	ServerVersion() string
	OpenAPIVersion() string
	CallAPI(string, interface{}, map[string]string, ...map[string]string) ([]byte, error)
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
}

type client struct {
	httpClient

	openAPI *openAPIInfo
	server  string
	token   string
}

func (c *client) Init() error {
	c.httpClient = new(http.Client)
	return nil
}

func (c *client) SetServer(url string) {
	c.server = url
}

func (c *client) SetToken(token string) {
	c.token = token
}

func (c *client) LoadSpec() error {
	resp, err := c.Get(strings.TrimSuffix(c.server, "/v1") + "/docs/openapi.json")
	if err != nil {
		return errors.Annotate(err, "get openapi spec")
	}
	var bytes []byte
	if resp.Body != nil {
		bytes, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return errors.Trace(err)
		}
	}
	if resp.StatusCode >= http.StatusMultipleChoices {
		return errors.Errorf("status: %s, body: %s", resp.Status, string(bytes))
	}

	if err = c.ParseOpenAPISpec(bytes); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *client) ParseOpenAPISpec(bytes []byte) error {
	c.openAPI = new(openAPIInfo)
	if err := json.Unmarshal(bytes, c.openAPI); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (c *client) ServerVersion() string {
	return c.openAPI.Version
}

func (c *client) OpenAPIVersion() string {
	return c.openAPI.OpenAPI
}

func (c *client) CallAPI(operationID string, body interface{}, pathParams map[string]string,
	queryParams ...map[string]string) ([]byte, error) {

	if c.openAPI == nil {
		return nil, nil
	}
	methodInfo, ok := c.openAPI.Paths.OperationIDs[operationID]
	if !ok {
		return nil, errors.Errorf("operation id %s not found", operationID)
	}
	reqPath := c.server + methodInfo.URL
	for _, param := range methodInfo.PathParams {
		val, ok := pathParams[param]
		if !ok {
			return nil, errors.Errorf("path param %s not set", param)
		}
		reqPath = strings.Replace(reqPath, "{"+param+"}", val, -1)
	}
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Trace(err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequest(methodInfo.Method, reqPath, bodyReader)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if c.token != "" {
		req.Header.Add("Xms-Auth-Token", c.token)
	}
	if len(queryParams) != 0 {
		q := req.URL.Query()
		for key, val := range queryParams[0] {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var bytes []byte
	if resp.Body != nil {
		bytes, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if resp.StatusCode >= 300 {
		return nil, errors.Errorf("status: %s, body: %s", resp.Status, string(bytes))
	}

	return bytes, nil
}

// NewOpenAPIClient returns a openapi client instance
func NewOpenAPIClient() Client {
	return &client{}
}
