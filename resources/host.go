package formation

import (
	"math/rand"
	"time"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// HostCreateReq defines create host request
type HostCreateReq struct {
	Host struct {

		// admin ip
		AdminIP string `json:"admin_ip"`

		// host description
		Description string `json:"description,omitempty"`

		// gateway ips for s3
		GatewayIps []string `json:"gateway_ips,omitempty"`

		// meta device for docker
		MetaDevice string `json:"meta_device,omitempty"`

		// cluster private ip for internal data access
		PrivateIP string `json:"private_ip,omitempty"`

		// deprecated
		ProtectionDomainID int64 `json:"protection_domain_id,omitempty"`

		// public ip for outside data access
		PublicIP string `json:"public_ip,omitempty"`

		// host roles: admin,monitor,block_storage_gateway,file_storage_gateway,s3_gateway,nfs_gateway
		Roles []string `json:"roles,omitempty"`

		// storage server or storage client
		Type string `json:"type,omitempty"`
	} `json:"host"`
}

// Host resource
type Host struct {
	ResourceBase
	AdminIP            *parser.StringExpr
	ProtectionDomainID *parser.IntegerExpr
	Description        *parser.StringExpr
	Roles              *parser.StringListExpr
	Type               *parser.StringExpr
}

// Init inits resource instance
func (host *Host) Init(stack utils.StackInterface) {
	host.ResourceBase.Init(stack)
	host.setDelegate(host)
}

// GetType return resource type
func (host *Host) GetType() string {
	return utils.ResourceHost
}

// CheckInterval return CheckInterval
func (host *Host) CheckInterval() int {
	return 15
}

// IsReady check if the formation args are ready
func (host *Host) IsReady() (ready bool) {
	if !host.isReady(host.AdminIP) ||
		!host.isReady(host.ProtectionDomainID) ||
		!host.isReady(host.Description) ||
		!host.isReady(host.Roles) {
		return false
	}

	return true
}

func (host *Host) fakeCreate() (bool, error) {
	host.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (host *Host) Create() (created bool, err error) {
	if config.DryRun {
		return host.fakeCreate()
	}

	req := new(HostCreateReq)
	if host.AdminIP != nil {
		req.Host.AdminIP = host.getStringValue(host.AdminIP)
	}
	if host.ProtectionDomainID != nil {
		req.Host.ProtectionDomainID = host.getIntegerValue(host.ProtectionDomainID)
	}
	if host.Description != nil {
		req.Host.Description = host.getStringValue(host.Description)
	}
	if host.Roles != nil {
		req.Host.Roles = host.getStringListValue(host.Roles)
	}
	if host.Type != nil {
		req.Host.Type = host.getStringValue(host.Type)
	}

	body, err := host.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create host with admin ip %s", host.AdminIP)
	}
	id, _, err := host.getIdentifyAndStatus(body)
	host.repr = id

	return false, nil
}

// IsCreated if the resource has been created
func (host *Host) IsCreated() (created bool, err error) {
	created, err = host.ResourceBase.IsCreated()
	if err != nil {
		return false, errors.Trace(err)
	}
	if created {
		time.Sleep(5 * time.Second)
	}
	return created, nil
}
