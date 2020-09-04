package formation

import (
	"bufio"
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	openapiClient "xsky.com/sds-formation/openapi-client"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// Resource resource
type Resource struct {
	utils.ResourceInterface
	Name string
}

// Stack stack
type Stack struct {
	token            string
	openapiClient    openapiClient.Client
	resourceValueMap map[string]interface{}
	templateContext  map[string]interface{}
	valueContexts    list.List
	template         *Template
	inTmpl           bool
	cacheIndex       int
	cacheExprs       []*CacheRecord
	cacheFile        io.ReadWriteCloser
	cacheFilePath    string
}

func (s *Stack) loadCache(name string) error {
	if err := Mkdir(config.CachePath, 0755); err != nil {
		return errors.Trace(err)
	}
	s.cacheFilePath = filepath.Join(config.CachePath, name)
	openMode := os.O_RDWR | os.O_CREATE | os.O_SYNC
	if config.NoContinue {
		openMode |= os.O_TRUNC
	}
	cacheFile, err := OpenFile(s.cacheFilePath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		return errors.Trace(err)
	}
	s.cacheFile = cacheFile
	cacheData, err := ioutil.ReadAll(s.cacheFile)
	if err != nil {
		return errors.Trace(err)
	}
	reader := bufio.NewReader(bytes.NewReader(cacheData))
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Trace(err)
		}
		cacheRecord := new(CacheRecord)
		if err = json.Unmarshal(line, cacheRecord); err != nil {
			return errors.Trace(err)
		}
		s.cacheExprs = append(s.cacheExprs, cacheRecord)
	}
	if len(s.cacheExprs) != 0 {
		log.Printf("Load %d resource cache record(s) from %s\n", len(s.cacheExprs), s.cacheFilePath)
	}
	return nil
}

// Init initialize the stack
func (s *Stack) Init(filePath string) (err error) {
	s.resourceValueMap = make(map[string]interface{})
	s.template = new(Template)

	file, err := OpenFile(filePath, os.O_RDONLY)
	if err != nil {
		return errors.Trace(err)
	}
	out, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.Trace(err)
	}
	if err = file.Close(); err != nil {
		return errors.Trace(err)
	}
	err = json.Unmarshal(out, s.template)
	if err != nil {
		return errors.Trace(err)
	}
	err = s.template.CheckTemplates()
	if err != nil {
		return errors.Trace(err)
	}

	var clusterURL string
	// TODO(wuhao): currently we only support using default as value of parameter.
	for key, param := range s.template.Parameters {
		ok := true
		switch key {
		case utils.ParamClusterURL:
			clusterURL, ok = param.Value.(string)
		default:
			s.resourceValueMap[key] = param.Value
		}
		if !ok {
			log.Fatalf("invalid input %s", key)
		}
	}
	if clusterURL == "" {
		log.Fatalf("%s is required", utils.ParamClusterURL)
	}

	templateHash, err := utils.GetHashString([]byte(s.template.Description + clusterURL))
	if err != nil {
		return errors.Trace(err)
	}
	if err = s.loadCache(templateHash); err != nil {
		return errors.Trace(err)
	}

	s.openapiClient = openapiClient.NewOpenAPIClient()
	s.openapiClient.SetServer(clusterURL)
	s.openapiClient.SetToken(config.Token)
	s.openapiClient.Init()
	if err = s.openapiClient.LoadSpec(); err != nil {
		return errors.Trace(err)
	}

	s.token = config.Token
	for _, r := range s.template.Resources {
		r.Properties.Init(s)
	}

	return
}

// CallAPI calls openapi api with operation id
func (s *Stack) CallAPI(api string, req interface{}, pathParam map[string]string,
	urlParam ...map[string]string) ([]byte, error) {

	bytes, err := s.GetOpenAPIClient().CallAPI(api, req, pathParam, urlParam...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return bytes, nil
}

// GetOpenAPIClient return api client
func (s *Stack) GetOpenAPIClient() openapiClient.Client {
	return s.openapiClient
}

func (s *Stack) getCreatingResources() []string {
	creatingResources := []string{}
	cacheIndex := 0
	for _, r := range s.template.Resources {
		name := r.Name
		for cacheIndex < len(s.cacheExprs) && s.cacheExprs[cacheIndex].InTemplate {
			cacheIndex++
		}
		if cacheIndex < len(s.cacheExprs) && name == r.Name && r.Type != utils.ResourceToken {
			name += "(cached)"
		}
		if r.Type != utils.ResourceToken {
			cacheIndex++
		}
		creatingResources = append(creatingResources, name)
	}
	return creatingResources
}

// CheckGetTemplateResourceContext make context list for template resource
func (s *Stack) CheckGetTemplateResourceContext(resource *ResourceInTemplate) ([]map[string]interface{}, error) {
	var contextList []map[string]interface{}
	valueMap := map[string]interface{}{}
	var makeContextSlice func([]*templateContext, map[string]interface{}) error
	makeContextSlice = func(contexts []*templateContext, values map[string]interface{}) error {
		if len(contexts) == 0 {
			if len(values) != 0 {
				newContext := make(map[string]interface{})
				for k, v := range values {
					newContext[k] = v
				}
				contextList = append(contextList, newContext)
			}
			return nil
		}
		context := contexts[0]
		if !context.Value.IsReady(s) {
			return errors.Errorf("template resource %s's context %s not found",
				resource.Name, context.Name)
		}
		var err error
		var value interface{}
		var rangeType bool
		switch val := context.Value.(type) {
		case *parser.IntegerExpr:
			value, err = val.GetValue(s)
			if err != nil {
				return errors.Trace(err)
			}
		case *parser.StringExpr:
			value, err = val.GetValue(s)
			if err != nil {
				return errors.Trace(err)
			}
		case *parser.BoolExpr:
			value, err = val.GetValue(s)
			if err != nil {
				return errors.Trace(err)
			}
		case *parser.StringListExpr:
			value, err = val.GetValue(s)
			if err != nil {
				return errors.Trace(err)
			}
			if context.Action == utils.ContextValueActionRange {
				rangeType = true
			}
		case *parser.IntegerListExpr:
			value, err = val.GetValue(s)
			if err != nil {
				return errors.Trace(err)
			}
			if context.Action == utils.ContextValueActionRange {
				rangeType = true
			}
		}
		if rangeType {
			contextValue := reflect.ValueOf(value)
			if contextValue.Kind() != reflect.Slice {
				return errors.Errorf("kind of context %s should be slice", context.Name)
			}
			contextValueLen := contextValue.Len()
			for i := 0; i < contextValueLen; i++ {
				valElem := contextValue.Index(i).Interface()
				valueMap[context.Name] = valElem
				if err = makeContextSlice(contexts[1:], valueMap); err != nil {
					return errors.Trace(err)
				}
				delete(valueMap, context.Name)
			}
		} else {
			valueMap[context.Name] = value
			if err = makeContextSlice(contexts[1:], valueMap); err != nil {
				return errors.Trace(err)
			}
			delete(valueMap, context.Name)
		}
		return nil
	}
	if err := makeContextSlice(resource.Context, valueMap); err != nil {
		return nil, errors.Trace(err)
	}

	// if no context settings for template resource, app a empty context to context list so that
	//  tempate run at least once(run template without context, this could happend).
	if len(contextList) == 0 {
		contextList = append(contextList, make(map[string]interface{}))
	}

	return contextList, nil
}

func (s *Stack) restoreCache(resource *ResourceInTemplate) (bool, error) {
	// do not cache token record and do not restore token cache record
	if s.cacheIndex >= len(s.cacheExprs) || resource.Type == utils.ResourceToken {
		return false, nil
	}
	cacheExpr := s.cacheExprs[s.cacheIndex]
	if resource.Name != cacheExpr.Name || resource.Type != cacheExpr.ResourceType {
		return false, errors.Errorf("got invalid cache record %s for resource %s, index %d",
			cacheExpr, resource.Name, s.cacheIndex)
	}
	cacheVal, err := cacheExpr.GetExpr()
	if err != nil {
		return false, errors.Trace(err)
	}
	s.resourceValueMap[resource.Name] = cacheVal
	s.cacheIndex++
	return true, nil
}

func (s *Stack) createResourcesWithTemplate(r *ResourceInTemplate) (interface{}, bool, error) {
	s.inTmpl = true
	defer func() {
		s.inTmpl = false
	}()
	templateContextes, err := s.CheckGetTemplateResourceContext(r)
	if err != nil {
		return nil, false, errors.Trace(err)
	}
	log.SetPrefix(fmt.Sprintf("[template resource: %s]+", r.Name))
	defer func() {
		log.SetPrefix("")
	}()
	templateValues := make([]map[string]interface{}, 0, len(templateContextes))
	for _, context := range templateContextes {
		s.pushContext(s.resourceValueMap)
		s.templateContext = context
		s.resourceValueMap = map[string]interface{}{}
		templateData := s.getTemplate(r.TemplateName)
		if templateData == nil {
			return nil, false, errors.Errorf("tempalte with name %s not found", r.TemplateName)
		}
		var resources []*ResourceInTemplate
		if err := json.Unmarshal(templateData, &resources); err != nil {
			return nil, false, errors.Annotatef(err, "parse template %s", r.TemplateName)
		}
		for _, resource := range resources {
			resource.Properties.Init(s)
		}
		if err := s.CreateResources(resources); err != nil {
			return nil, false, errors.Trace(err)
		}
		templateValues = append(templateValues, s.resourceValueMap)
		s.resourceValueMap = s.popContext()
		s.templateContext = nil
	}
	restored, err := s.restoreCache(r)
	if err != nil {
		return nil, false, errors.Trace(err)
	}
	if restored {
		return nil, true, nil
	}
	s.resourceValueMap[r.Name] = templateValues
	return templateValues, false, nil
}

// CreateResources create resources with Resource Template
func (s *Stack) CreateResources(resources []*ResourceInTemplate) error {
	for _, r := range resources {
		var repr interface{}
		var rType string
		if r.Type == utils.ResourceTemplate {
			tmplRepr, restored, err := s.createResourcesWithTemplate(r)
			if err != nil {
				return errors.Trace(err)
			}
			if restored {
				continue
			}
			repr = tmplRepr
			rType = utils.ResourceTemplate
			log.Printf("template resource %s has been created successfully!!!", r.Name)
		} else {
			restored, err := s.restoreCache(r)
			if err != nil {
				return errors.Trace(err)
			}
			if restored {
				continue
			}
			name, resource := r.Name, r.Properties
			if !resource.IsReady() {
				log.Fatalf("resources %v can't be created for lack of required resources", name)
			}

			switch r.Action {
			case utils.ActionTypeUpdate:
				err := s.handleUpdate(name, resource, r.WaitInterval, r.CheckInterval)
				if err != nil {
					log.Fatalf("update resource %s: %s", name, errors.ErrorStack(err))
				}
			case utils.ActionTypeGet:
				err := s.handleGet(name, resource, r.WaitInterval, r.CheckInterval)
				if err != nil {
					log.Fatalf("get resource %s: %s", name, errors.ErrorStack(err))
				}
				s.resourceValueMap[name] = resource.Repr()
			default:
				err := s.handleCreate(name, resource, r.WaitInterval, r.CheckInterval)
				if err != nil {
					log.Fatalf("create resource %s: %s", name, errors.ErrorStack(err))
				}
				s.resourceValueMap[name] = resource.Repr()
			}
			if r.Sleep > 0 {
				log.Printf("sleep %d seconds", r.Sleep)
				time.Sleep(time.Duration(r.Sleep) * time.Second)
			}
			repr = r.Properties.Repr()
			rType = r.Properties.GetType()
		}
		if err := s.record(r.Name, rType, repr); err != nil {
			log.Fatal(errors.ErrorStack(err))
		}
	}

	return nil
}

// Create create resource in the stack
func (s *Stack) Create() {
	log.Printf("Stack started create resources: %+v", s.getCreatingResources())

	if err := s.CreateResources(s.template.Resources); err != nil {
		log.Fatal(err)
	}

	if e := s.cacheFile.Close(); e != nil {
		log.Println(errors.Annotate(e, "close cache file"))
	}
	if e := os.Remove(s.cacheFilePath); e != nil {
		log.Println(errors.Annotate(e, "remove cache file"))
	}

	return
}

func (s *Stack) record(resourceName, resourceType string, value interface{}) error {
	if resourceType == utils.ResourceToken {
		return nil
	}
	cacheRecord, err := GetCacheRecord(resourceName, resourceType, value, s.inTmpl)
	if err != nil {
		return errors.Trace(err)
	}
	bytes, err := json.Marshal(cacheRecord)
	if err != nil {
		return errors.Trace(err)
	}
	_, err = s.cacheFile.Write(append(bytes, '\n'))
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *Stack) handleGet(
	name string, resource utils.ResourceInterface, waitInterval, checkInterval int) (err error) {

	rType := resource.GetType()
	log.Printf("try to update resource %s of type %s...", name, rType)

	err = resource.Get()
	if err != nil {
		return errors.Annotatef(err, "get resource %s of type %s", name, rType)
	}

	log.Printf("got resource %s with representation %v successfully!!!", name, resource.Repr())
	return nil
}

func (s *Stack) handleUpdate(
	name string, resource utils.ResourceInterface, waitInterval, checkInterval int) (err error) {

	repr, ok := s.resourceValueMap[name]
	if !ok {
		return errors.Errorf("failed to get resource %s", name)
	}
	rType := resource.GetType()
	log.Printf("try to update resource %s of type %s...", name, rType)

	updated, err := resource.Update(repr)
	if err != nil {
		return errors.Annotatef(err, "failed to update resource %s of type %s", name, rType)
	}

	if !updated {
		if err = s.waitUpdated(name, resource, waitInterval, checkInterval); err != nil {
			return errors.Trace(err)
		}
	}

	log.Printf("resource %s with representation %v has been updated successfully!!!",
		name, resource.Repr())
	return nil
}

func (s *Stack) waitUpdated(
	name string, resource utils.ResourceInterface, waitInterval, checkInterval int) (err error) {

	if waitInterval > 0 {
		time.Sleep(time.Duration(waitInterval) * time.Second)
	}

	log.Printf("start to check status of resource %s", name)
	for i := 1; i <= utils.DefaultCheckCount; i++ {
		log.Printf("check %d time(s).", i)
		updated, err := resource.IsUpdated()
		if err != nil {
			return errors.Trace(err)
		}
		if updated {
			return nil
		}

		if checkInterval > 0 {
			time.Sleep(time.Duration(checkInterval) * time.Second)
		} else {
			time.Sleep(time.Duration(resource.CheckInterval()) * time.Second)
		}
	}

	return errors.Errorf("timeout for waiting resource %s to be updated", name)
}

func (s *Stack) handleCreate(
	name string, resource utils.ResourceInterface, waitInterval, checkInterval int) (err error) {

	rType := resource.GetType()
	log.Printf("try to create resource %s of type %s...", name, rType)
	if rType != utils.ResourceToken && s.token == "" {
		return errors.Errorf("create resource %s without token", name)
	}

	created, err := resource.Create()
	if err != nil {
		return errors.Annotatef(err, "failed to create resource %s of type %s", name, rType)
	}
	if !created {
		if err = s.waitCreated(name, resource, waitInterval, checkInterval); err != nil {
			return errors.Trace(err)
		}
	}

	log.Printf("resource %s with representation %v has been created successfully!!!",
		name, resource.Repr())
	if rType == utils.ResourceToken {
		s.token = resource.Repr().(string)
		s.GetOpenAPIClient().SetToken(s.token)
		log.Printf("reset %s to %s", utils.XmsHeaderAuthToken, s.token)
	}
	return nil
}

func (s *Stack) waitCreated(
	name string, resource utils.ResourceInterface, waitInterval, checkInterval int) (err error) {

	if waitInterval > 0 {
		time.Sleep(time.Duration(waitInterval) * time.Second)
	}

	log.Printf("start to check status of resource %s", name)
	for i := 1; i <= utils.DefaultCheckCount; i++ {
		log.Printf("check %d time(s).", i)
		created, err := resource.IsCreated()
		if err != nil {
			return errors.Trace(err)
		}
		if created {
			return nil
		}

		if checkInterval > 0 {
			time.Sleep(time.Duration(checkInterval) * time.Second)
		} else {
			time.Sleep(time.Duration(resource.CheckInterval()) * time.Second)
		}
	}

	return errors.Errorf("timeout for waiting resource %s to be created", name)
}

// GetResourceValue returns resource value with specific name
// value search order:
//      1. template context
//      2. resource value map
//      3. value contexts
func (s *Stack) GetResourceValue(name string) interface{} {
	// TODO: return as val, exist format
	if s.templateContext != nil {
		value := s.templateContext[name]
		if value != nil {
			return value
		}
	}
	value := s.resourceValueMap[name]
	if value != nil {
		return value
	}
	contextBack := s.valueContexts.Back()
	for contextBack != nil {
		values := contextBack.Value.(map[string]interface{})
		value := values[name]
		if value != nil {
			return value
		}
		contextBack = contextBack.Prev()
	}
	return nil
}

func (s *Stack) pushContext(context map[string]interface{}) {
	s.valueContexts.PushBack(context)
}

func (s *Stack) popContext() map[string]interface{} {
	last := s.valueContexts.Back()
	if last == nil {
		return nil
	}
	s.valueContexts.Remove(last)
	return last.Value.(map[string]interface{})
}

func (s *Stack) getTemplate(name string) json.RawMessage {
	if s.template.Templates == nil {
		return nil
	}
	return s.template.Templates[name]
}
