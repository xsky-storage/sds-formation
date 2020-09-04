package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// OSGatewayCreateReq defines os gateway create request
type OSGatewayCreateReq struct {
	OSGateway struct {
		Description string `json:"description,omitempty"`

		GatewayIP string `json:"gateway_ip,omitempty"`

		HostID int64 `json:"host_id"`

		Name string `json:"name"`

		Port int64 `json:"port"`
	} `json:"os_gateway"`
}

// ObjectStorageGateway resource
type ObjectStorageGateway struct {
	ResourceBase

	Description *parser.StringExpr
	GatewayIP   *parser.StringExpr
	HostID      *parser.IntegerExpr
	Name        *parser.StringExpr
	Port        *parser.IntegerExpr
}

// Init inits resource instance
func (gateway *ObjectStorageGateway) Init(stack utils.StackInterface) {
	gateway.ResourceBase.Init(stack)
	gateway.setDelegate(gateway)
}

// GetType return resource type
func (gateway *ObjectStorageGateway) GetType() string {
	return utils.ResourceObjectStorageGateway
}

// IsReady check if the formation args are ready
func (gateway *ObjectStorageGateway) IsReady() (ready bool) {
	if !gateway.isReady(gateway.Description) ||
		!gateway.isReady(gateway.GatewayIP) ||
		!gateway.isReady(gateway.HostID) ||
		!gateway.isReady(gateway.Name) ||
		!gateway.isReady(gateway.Port) {
		return false
	}

	return true
}

func (gateway *ObjectStorageGateway) fakeCreate() (bool, error) {
	gateway.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (gateway *ObjectStorageGateway) Create() (created bool, err error) {
	if gateway.Name == nil {
		err = errors.Errorf("Name is required for resource %s", gateway.GetType())
		return
	}
	if config.DryRun {
		return gateway.fakeCreate()
	}

	name := gateway.getStringValue(gateway.Name)
	resourceID, err := gateway.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		gateway.repr = resourceID
		return false, nil
	}

	req := new(OSGatewayCreateReq)
	gatewayInfo := &req.OSGateway
	if gateway.Description != nil {
		gatewayInfo.Description = gateway.getStringValue(gateway.Description)
	}
	if gateway.GatewayIP != nil {
		gatewayInfo.GatewayIP = gateway.getStringValue(gateway.GatewayIP)
	}
	if gateway.Name != nil {
		gatewayInfo.Name = gateway.getStringValue(gateway.Name)
	}
	if gateway.HostID != nil {
		gatewayInfo.HostID = gateway.getIntegerValue(gateway.HostID)
	}
	if gateway.Port != nil {
		gatewayInfo.Port = gateway.getIntegerValue(gateway.Port)
	}

	body, err := gateway.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "failed to create object storage gateway %s", name)
	}

	id, _, err := gateway.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	gateway.repr = id
	return false, nil
}
