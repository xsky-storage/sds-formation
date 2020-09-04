package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type fsGatewayReq struct {
	// host id
	HostID int64 `json:"host_id" required:"true"`
	// network address id
	NetworkAddressID int64 `json:"network_address_id"`
}

// FSGatewayGroupCreateReq defines request for creating gateway group
type FSGatewayGroupCreateReq struct {
	GatewayGroup struct {
		// name of gateway group
		Name string `json:"name" required:"true"`
		// description of gateway group
		Description string `json:"description,omitempty"`
		// file storage gateways list
		Gateways []*fsGatewayReq `json:"fs_gateways" required:"true"`
		// virtual ip of gateway group
		VIP string `json:"vip" required:"true"`
		// types of supported (smb, nfs, ftp)
		Types []string `json:"types" required:"true"`
		// smb security type
		Security *string `json:"security,omitempty"`
		// smb version 1.0 enabled
		SMB1Enabled *bool `json:"smb1_enabled,omitempty"`
		// smb ports
		SMBPorts []int64 `json:"smb_ports,omitempty"`
		// nfs versions of nfs supported
		NFSVersions []string `json:"nfs_versions,omitempty"`
		// ftp encoding format, default is utf8
		Encoding *string `json:"encoding,omitempty"`
	} `json:"fs_gateway_group" required:"true"`
}

type fsGateway struct {
	HostID           *parser.IntegerExpr
	NetworkAddressID *parser.IntegerExpr
}

// FSGatewayGroup resource
type FSGatewayGroup struct {
	ResourceBase

	Name        *parser.StringExpr
	Description *parser.StringExpr
	VIP         *parser.StringExpr
	Types       *parser.StringListExpr
	Security    *parser.StringExpr
	SMB1Enabled *parser.BoolExpr
	SMBPorts    *parser.IntegerListExpr
	NFSVersions *parser.StringListExpr
	Encoding    *parser.StringExpr
	Gateways    []*fsGateway
}

// Init inits resource instance
func (gatewayGroup *FSGatewayGroup) Init(stack utils.StackInterface) {
	gatewayGroup.ResourceBase.Init(stack)
	gatewayGroup.setDelegate(gatewayGroup)
}

// GetType return resource type
func (gatewayGroup *FSGatewayGroup) GetType() string {
	return utils.ResourceFSGatewayGroup
}

// IsReady check if the formation args are ready
func (gatewayGroup *FSGatewayGroup) IsReady() (ready bool) {
	if !gatewayGroup.isReady(gatewayGroup.Name) ||
		!gatewayGroup.isReady(gatewayGroup.Description) ||
		!gatewayGroup.isReady(gatewayGroup.VIP) ||
		!gatewayGroup.isReady(gatewayGroup.Types) ||
		!gatewayGroup.isReady(gatewayGroup.Security) ||
		!gatewayGroup.isReady(gatewayGroup.SMB1Enabled) ||
		!gatewayGroup.isReady(gatewayGroup.SMBPorts) ||
		!gatewayGroup.isReady(gatewayGroup.NFSVersions) ||
		!gatewayGroup.isReady(gatewayGroup.Encoding) {
		return false
	}
	for _, gateway := range gatewayGroup.Gateways {
		if !gatewayGroup.isReady(gateway.HostID) ||
			!gatewayGroup.isReady(gateway.NetworkAddressID) {
			return false
		}
	}

	return true
}

func (gatewayGroup *FSGatewayGroup) fakeCreate() (bool, error) {
	gatewayGroup.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (gatewayGroup *FSGatewayGroup) Create() (created bool, err error) {
	if gatewayGroup.Name == nil {
		err = errors.Errorf("Name is required for resource %s", gatewayGroup.GetType())
		return
	}
	if config.DryRun {
		return gatewayGroup.fakeCreate()
	}

	name := gatewayGroup.getStringValue(gatewayGroup.Name)
	resourceID, err := gatewayGroup.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		gatewayGroup.repr = resourceID
		return false, nil
	}

	req := new(FSGatewayGroupCreateReq)
	groupInfo := &req.GatewayGroup
	if gatewayGroup.Name != nil {
		groupInfo.Name = gatewayGroup.getStringValue(gatewayGroup.Name)
	}
	if gatewayGroup.Description != nil {
		groupInfo.Description = gatewayGroup.getStringValue(gatewayGroup.Description)
	}
	if gatewayGroup.VIP != nil {
		groupInfo.VIP = gatewayGroup.getStringValue(gatewayGroup.VIP)
	}
	if gatewayGroup.Types != nil {
		groupInfo.Types = gatewayGroup.getStringListValue(gatewayGroup.Types)
	}
	if gatewayGroup.Security != nil {
		security := gatewayGroup.getStringValue(gatewayGroup.Security)
		groupInfo.Security = &security
	}
	if gatewayGroup.SMB1Enabled != nil {
		smb1Enabled := gatewayGroup.getBoolValue(gatewayGroup.SMB1Enabled)
		groupInfo.SMB1Enabled = &smb1Enabled
	}
	if gatewayGroup.SMBPorts != nil {
		groupInfo.SMBPorts = gatewayGroup.getIntegerListValue(gatewayGroup.SMBPorts)
	}
	if gatewayGroup.NFSVersions != nil {
		groupInfo.NFSVersions = gatewayGroup.getStringListValue(gatewayGroup.NFSVersions)
	}
	if gatewayGroup.Encoding != nil {
		encoding := gatewayGroup.getStringValue(gatewayGroup.Encoding)
		groupInfo.Encoding = &encoding
	}
	for _, gateway := range gatewayGroup.Gateways {
		gatewayReq := new(fsGatewayReq)
		if gateway.HostID != nil {
			gatewayReq.HostID = gatewayGroup.getIntegerValue(gateway.HostID)
		}
		if gateway.NetworkAddressID != nil {
			gatewayReq.NetworkAddressID = gatewayGroup.getIntegerValue(gateway.NetworkAddressID)
		}
		groupInfo.Gateways = append(groupInfo.Gateways, gatewayReq)
	}

	body, err := gatewayGroup.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs gateway group %s", name)
	}

	id, status, err := gatewayGroup.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	gatewayGroup.repr = id
	created, err = gatewayGroup.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
