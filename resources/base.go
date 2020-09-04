package formation

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/juju/errors"

	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type commonInstance struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	ActionStatus string `json:"action_status"`
}

// ResourceBase basic of all resources
type ResourceBase struct {
	repr     interface{}
	stack    utils.StackInterface
	delegate utils.ResourceInterface

	recordInstance reflect.Type
}

// CallResourceAPI call resource api
func (r *ResourceBase) CallResourceAPI(apiType string, req interface{}, pathParam map[string]string,
	queryParam ...map[string]string) ([]byte, error) {

	api, err := settings.GetSetting(r.GetType(), apiType)
	if err != nil {
		return nil, errors.Trace(err)
	}
	body, err := r.stack.CallAPI(api, req, pathParam, queryParam...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return body, nil
}

// CallGetAPI calls get api of resource instance
func (r *ResourceBase) CallGetAPI(pathParams ...map[string]string) ([]byte, error) {
	pathParam := map[string]string{}
	// get req identify key could not exist
	getReqIdentify, _ := settings.GetSetting(r.GetType(), utils.GetReqIdentify)
	if getReqIdentify != "" && r.repr != nil && reflect.TypeOf(r.repr).Kind() <= reflect.Float64 {
		valStr, err := r.getValString(r.repr)
		if err != nil {
			return nil, errors.Annotatef(err, "parse resource repr to string")
		}
		pathParam[getReqIdentify] = valStr
	}
	if len(pathParams) != 0 {
		for key, val := range pathParams[0] {
			pathParam[key] = val
		}
	}

	body, err := r.CallResourceAPI(utils.GetAPIName, nil, pathParam)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return body, nil
}

// CallCreateAPI call create api of resource
func (r *ResourceBase) CallCreateAPI(req interface{}, pathParam map[string]string, queryParam ...map[string]string) ([]byte, error) {
	body, err := r.CallResourceAPI(utils.CreateAPIName, req, pathParam, queryParam...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return body, nil
}

func (r *ResourceBase) isReady(expr parser.ExprType) (ready bool) {
	if expr == nil || reflect.ValueOf(expr).IsNil() {
		return true
	}
	return expr.IsReady(r.stack)
}

func (r *ResourceBase) getResourceByName(name string, queryParams ...map[string]string) (interface{}, error) {
	id, err := r.getResourceFromListAPI("Name", name, queryParams...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return id, nil
}

func (r *ResourceBase) getResourceFromListAPI(field string, val interface{},
	queryParams ...map[string]string) (resourceID interface{}, err error) {

	if len(queryParams) == 0 {
		queryParams = append(queryParams, map[string]string{})
	}
	queryParams[0]["limit"] = "-1"

	apiName, err := settings.GetSetting(r.GetType(), utils.ListAPIName)
	if err != nil {
		return nil, errors.Trace(err)
	}
	body, err := r.stack.GetOpenAPIClient().CallAPI(apiName, nil, nil, queryParams...)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to list resource")
	}

	recordsMap := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &recordsMap); err != nil {
		return nil, errors.Trace(err)
	}
	recordsKey, err := settings.GetSetting(r.GetType(), utils.RecordsKey)
	if err != nil {
		return nil, errors.Trace(err)
	}
	recordsData, ok := recordsMap[recordsKey]
	if !ok {
		return nil, errors.Errorf("key %s not found in records data", recordsKey)
	}
	instances := reflect.New(reflect.SliceOf(r.recordInstance)).Interface()
	if err = json.Unmarshal(recordsData, instances); err != nil {
		return nil, errors.Trace(err)
	}

	rawVal := reflect.Indirect(reflect.ValueOf(instances))
	if rawVal.Len() == 0 {
		return nil, nil
	}
	identifyKey := settings.GetIdentifyKey(r.GetType())
	if field == "" {
		return reflect.Indirect(rawVal.Index(0)).FieldByName(identifyKey).Interface(), nil
	}

	for i := 0; i < rawVal.Len(); i++ {
		instance := rawVal.Index(i)
		instanceVal := reflect.Indirect(instance).FieldByName(field).Interface()
		if reflect.DeepEqual(instanceVal, val) {
			return reflect.Indirect(instance.FieldByName(identifyKey)).Interface(), nil
		}
	}
	return nil, nil
}

func (r *ResourceBase) getStringValue(expr *parser.StringExpr) string {
	value, err := expr.GetValue(r.stack)
	if err != nil {
		log.Fatalf("failed to get string value: %s", errors.ErrorStack(err))
	}
	return value
}

func (r *ResourceBase) getIntegerValue(expr *parser.IntegerExpr) int64 {
	value, err := expr.GetValue(r.stack)
	if err != nil {
		log.Fatalf("failed to get integer value: %s", errors.ErrorStack(err))
	}
	return value
}

func (r *ResourceBase) getBoolValue(expr *parser.BoolExpr) bool {
	value, err := expr.GetValue(r.stack)
	if err != nil {
		log.Fatalf("failed to get bool value: %s", errors.ErrorStack(err))
	}
	return value
}

func (r *ResourceBase) getStringListValue(expr *parser.StringListExpr) []string {
	value, err := expr.GetValue(r.stack)
	if err != nil {
		log.Fatalf("failed to get string list value: %s", errors.ErrorStack(err))
	}
	return value
}

func (r *ResourceBase) getIntegerListValue(expr *parser.IntegerListExpr) []int64 {
	value, err := expr.GetValue(r.stack)
	if err != nil {
		log.Fatalf("failed to get integer list value: %s", errors.ErrorStack(err))
	}
	return value
}

func (r *ResourceBase) checkStatus(status string) (created bool, err error) {
	if status == utils.StatusActive || status == utils.StatusFinished || status == utils.StatusHealthy {
		created = true
	} else if status == utils.StatusError || status == utils.StatusVerifyingError ||
		status == utils.StatusSyncingError {

		err = errors.Errorf("resource is in status %s", status)
	}
	return
}

func (r *ResourceBase) checkResp(resp *http.Response) (err error) {
	if resp.StatusCode >= 400 {
		err = errors.Errorf("request failed with status code %d", resp.StatusCode)
		return
	}
	return
}

// Init initialize a resource
func (r *ResourceBase) Init(stack utils.StackInterface) {
	r.stack = stack

	r.recordInstance = reflect.TypeOf(commonInstance{})
}

func (r *ResourceBase) setDelegate(resource utils.ResourceInterface) {
	r.delegate = resource
}

// Repr get resource representation
func (r *ResourceBase) Repr() interface{} {
	return r.repr
}

// CheckInterval return check interval
func (r *ResourceBase) CheckInterval() int {
	return utils.DefaultCheckInterval
}

// IsReady check if resource is ready
func (r *ResourceBase) IsReady() bool {
	return false
}

// Get get the resource
func (r *ResourceBase) Get() (err error) {
	return errors.Errorf("Not implemented")
}

// Create create the resource
func (r *ResourceBase) Create() (created bool, err error) {
	return false, errors.Errorf("Not implemented")
}

// Update update a resource
func (r *ResourceBase) Update(repr interface{}) (updated bool, err error) {
	return false, errors.Errorf("Not implemented")
}

// IsUpdated check if a resource is updated
func (r *ResourceBase) IsUpdated() (updated bool, err error) {
	return false, errors.Errorf("Not implemented")
}

// Delete delete a resource
func (r *ResourceBase) Delete(repr interface{}) (deleted bool, err error) {
	return false, errors.Errorf("Not implemented")
}

// IsDeleted check if a resource is deleted
func (r *ResourceBase) IsDeleted() (deleted bool, err error) {
	return false, errors.Errorf("Not implemented")
}

// GetType calls GetType of real resource instance
func (r *ResourceBase) GetType() string {
	if r.delegate != nil {
		return r.delegate.GetType()
	}
	return ""
}

func (r *ResourceBase) getInstance(body []byte) (interface{}, error) {
	recordMap := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &recordMap); err != nil {
		return nil, errors.Annotatef(err, "parse response of ")
	}
	recordKey, err := settings.GetSetting(r.GetType(), utils.RecordKey)
	if err != nil {
		return nil, errors.Trace(err)
	}
	rawData, ok := recordMap[recordKey]
	if !ok {
		return nil, errors.Errorf("key %s not found in response data", recordKey)
	}
	instance := reflect.New(r.recordInstance).Interface()
	if err := json.Unmarshal(rawData, instance); err != nil {
		return nil, errors.Trace(err)
	}
	return instance, nil
}

func (r *ResourceBase) getIdentifyAndStatus(body []byte) (interface{}, string, error) {
	instance, err := r.getInstance(body)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	statusKey := settings.GetStatusKey(r.GetType())
	field := reflect.Indirect(reflect.ValueOf(instance)).FieldByName(statusKey)
	if !field.IsValid() {
		return nil, "", errors.Errorf("filed %s of %s not found", statusKey, r.GetType())
	}
	status := field.Interface().(string)

	identifyKey := settings.GetIdentifyKey(r.GetType())
	field = reflect.Indirect(reflect.ValueOf(instance)).FieldByName(identifyKey)
	if !field.IsValid() {
		return nil, "", errors.Errorf("filed %s of %s not found", identifyKey, r.GetType())
	}
	identify := field.Interface()
	return identify, status, nil
}

func (r *ResourceBase) getValString(val interface{}) (string, error) {
	typStr := reflect.TypeOf(val).String()
	switch {
	case strings.HasPrefix(typStr, "int"):
		return fmt.Sprintf("%d", val), nil
	case strings.HasPrefix(typStr, "uint"):
		return fmt.Sprintf("%d", val), nil
	case strings.HasPrefix(typStr, "float"):
		return fmt.Sprintf("%f", val), nil
	case strings.HasPrefix(typStr, "string"):
		return fmt.Sprintf("%s", val), nil
	case strings.HasPrefix(typStr, "bool"):
		return fmt.Sprintf("%t", val), nil
	}
	return "", errors.Errorf("unsupport value type: %s", typStr)
}

// IsCreated if the resource has been created
func (r *ResourceBase) IsCreated() (created bool, err error) {
	body, err := r.CallGetAPI(nil)
	if err != nil {
		return false, errors.Annotatef(err, "get resource with id %d", r.repr)
	}
	_, status, err := r.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	return r.checkStatus(status)
}

// NewResource returns a new resource object correspoding with the provided type
func NewResource(typeName string, action string) utils.ResourceInterface {
	switch typeName {
	case utils.ResourceIntegerList:
		return &IntegerList{}
	case utils.ResourceStringList:
		return &StringList{}
	case utils.ResourceHost:
		return &Host{}
	case utils.ResourceHosts:
		return &Hosts{}
	case utils.ResourceBlockVolume:
		return &BlockVolume{}
	case utils.ResourceBlockVolumes:
		return &BlockVolumes{}
	case utils.ResourceDiskList:
		if action == utils.ActionTypeUpdate {
			return &DiskListUpdate{}
		}
		return &DiskList{}
	case utils.ResourceOsd:
		return &Osd{}
	case utils.ResourceOsds:
		return &Osds{}
	case utils.ResourcePool:
		return &Pool{}
	case utils.ResourcePartitions:
		return &Partitions{}
	case utils.ResourceUser:
		return &User{}
	case utils.ResourceBootNode:
		return &BootNode{}
	case utils.ResourceObjectStorage:
		return &ObjectStorage{}
	case utils.ResourceObjectStorageUser:
		return &ObjectStorageUser{}
	case utils.ResourceObjectStoragePolicy:
		return &ObjectStoragePolicy{}
	case utils.ResourceObjectStorageBucket:
		return &ObjectStorageBucket{}
	case utils.ResourceObjectStorageGateway:
		return &ObjectStorageGateway{}
	case utils.ResourceNFSGateway:
		return &NFSGateway{}
	case utils.ResourceObjectStorageArchivePool:
		return &ObjectStorageArchivePool{}
	case utils.ResourceClientGroup:
		return &ClientGroup{}
	case utils.ResourceAccessPath:
		return &AccessPath{}
	case utils.ResourceMappingGroup:
		return &MappingGroup{}
	case utils.ResourceToken:
		return &Token{}
	case utils.ResourceS3LoadBalancerGroup:
		return &S3LoadBalancerGroup{}
	case utils.ResourceNetworkAddress:
		return &NetworkAddress{}
	case utils.ResourceFSAD:
		return &FSAD{}
	case utils.ResourceFSClient:
		return &FSClient{}
	case utils.ResourceFSClientGroup:
		return &FSClientGroup{}
	case utils.ResourceFSFTPShare:
		return &FSFTPShare{}
	case utils.ResourceFSFolder:
		return &FSFolder{}
	case utils.ResourceFSGatewayGroup:
		return &FSGatewayGroup{}
	case utils.ResourceFSLdap:
		return &FSLdap{}
	case utils.ResourceFSNFSShare:
		return &FSNFSShare{}
	case utils.ResourceFSQuotaTree:
		return &FSFolderQuotaTree{}
	case utils.ResourceFSSMBShare:
		return &FSSMBShare{}
	case utils.ResourceFSUser:
		return &FSUser{}
	case utils.ResourceFSUserGroup:
		return &FSUserGroup{}
	case utils.ResourceFSArbitrationPool:
		return &FSArbitrationPool{}
	case utils.ResourceTemplate:
		return &ResourceBase{}
	}

	return nil
}
