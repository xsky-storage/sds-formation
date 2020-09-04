package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// NFSGatewayCreateReq defines nfs gateway create request
type NFSGatewayCreateReq struct {
	NFSGateway struct {
		Description string `json:"description"`

		GatewayIP string `json:"gateway_ip,omitempty"`

		HostID int64 `json:"host_id"`

		Name string `json:"name"`

		Port int64 `json:"port"`
	} `json:"nfs_gateway"`
}

// NFSGateway resource
type NFSGateway struct {
	ResourceBase

	Description *parser.StringExpr
	GatewayIP   *parser.StringExpr
	HostID      *parser.IntegerExpr
	Name        *parser.StringExpr
	Port        *parser.IntegerExpr
}

// Init inits resource instance
func (gateway *NFSGateway) Init(stack utils.StackInterface) {
	gateway.ResourceBase.Init(stack)
	gateway.setDelegate(gateway)
}

// GetType return resource type
func (gateway *NFSGateway) GetType() string {
	return utils.ResourceNFSGateway
}

// IsReady check if the formation args are ready
func (gateway *NFSGateway) IsReady() (ready bool) {
	if !gateway.isReady(gateway.Description) ||
		!gateway.isReady(gateway.GatewayIP) ||
		!gateway.isReady(gateway.HostID) ||
		!gateway.isReady(gateway.Name) ||
		!gateway.isReady(gateway.Port) {
		return false
	}

	return true
}

func (gateway *NFSGateway) fakeCreate() (bool, error) {
	gateway.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (gateway *NFSGateway) Create() (created bool, err error) {
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
		return false, errors.Annotatef(err, "get nfs gatway %s", name)
	}
	if resourceID != nil {
		gateway.repr = resourceID
		return false, nil
	}

	req := new(NFSGatewayCreateReq)
	gatewayInfo := &req.NFSGateway
	if gateway.Description != nil {
		gatewayInfo.Description = gateway.getStringValue(gateway.Description)
	}
	if gateway.GatewayIP != nil {
		gatewayInfo.GatewayIP = gateway.getStringValue(gateway.GatewayIP)
	}
	if gateway.HostID != nil {
		gatewayInfo.HostID = gateway.getIntegerValue(gateway.HostID)
	}
	if gateway.Name != nil {
		gatewayInfo.Name = gateway.getStringValue(gateway.Name)
	}
	if gateway.Port != nil {
		gatewayInfo.Port = gateway.getIntegerValue(gateway.Port)
	}

	body, err := gateway.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create nfs gateway %s", name)
	}

	id, _, err := gateway.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	gateway.repr = id

	return false, nil
}
