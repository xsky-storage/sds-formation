package formation

import (
	"fmt"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// AccessPath resource
type AccessPath struct {
	ResourceBase

	Chap               *parser.BoolExpr
	Description        *parser.StringExpr
	HostIDs            *parser.IntegerListExpr
	MappingGroups      []*MappingGroup
	Name               *parser.StringExpr
	ProtectionDomainID *parser.IntegerExpr
	Tname              *parser.StringExpr
	Tsecret            *parser.StringExpr
	Type               *parser.StringExpr
}

// Init inits resource instance
func (accessPath *AccessPath) Init(stack utils.StackInterface) {
	accessPath.ResourceBase.Init(stack)
	accessPath.setDelegate(accessPath)
}

// GetType return resource type
func (accessPath *AccessPath) GetType() string {
	return utils.ResourceAccessPath
}

type mappingGroupReq struct {
	AccessPathID int64 `json:"access_path_id,omitempty"`

	BlockVolumeIds []int64 `json:"block_volume_ids,omitempty"`

	ClientGroupID int64 `json:"client_group_id,omitempty"`
}

// AccessPathCreateReq defines request for create access path
type AccessPathCreateReq struct {
	AccessPath struct {
		Chap bool `json:"chap,omitempty"`

		Description string `json:"description,omitempty"`

		HostIds []int64 `json:"host_ids,omitempty"`

		MappingGroups []mappingGroupReq `json:"mapping_groups,omitempty"`

		Name string `json:"name,omitempty"`

		ProtectionDomainID int64 `json:"protection_domain_id,omitempty"`

		Tname string `json:"tname,omitempty"`

		Tsecret string `json:"tsecret,omitempty"`

		Type string `json:"type,omitempty"`
	} `json:"access_path"`
}

// IsReady check if the formation args are ready
func (accessPath *AccessPath) IsReady() (ready bool) {
	if !accessPath.isReady(accessPath.Chap) ||
		!accessPath.isReady(accessPath.Description) ||
		!accessPath.isReady(accessPath.HostIDs) ||
		!accessPath.isReady(accessPath.Name) ||
		!accessPath.isReady(accessPath.ProtectionDomainID) ||
		!accessPath.isReady(accessPath.Tname) ||
		!accessPath.isReady(accessPath.Tsecret) ||
		!accessPath.isReady(accessPath.Type) {
		return false
	}
	for _, mappingGroup := range accessPath.MappingGroups {
		if !accessPath.isReady(mappingGroup.BlockVolumeIDs) ||
			!accessPath.isReady(mappingGroup.ClientGroupID) {
			return false
		}
	}

	return true
}

func (accessPath *AccessPath) fakeCreate() (bool, error) {
	accessPath.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (accessPath *AccessPath) Create() (created bool, err error) {
	if accessPath.Name == nil {
		err = fmt.Errorf("Name is required for resource %s", accessPath.GetType())
		return
	}
	if config.DryRun {
		return accessPath.fakeCreate()
	}

	name := accessPath.getStringValue(accessPath.Name)
	resourceID, err := accessPath.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		accessPath.repr = resourceID
		return false, nil
	}

	req := new(AccessPathCreateReq)
	req.AccessPath.Name = name
	if accessPath.Chap != nil {
		req.AccessPath.Chap = accessPath.getBoolValue(accessPath.Chap)
	}
	if accessPath.Description != nil {
		req.AccessPath.Description = accessPath.getStringValue(accessPath.Description)
	}
	if accessPath.HostIDs != nil {
		req.AccessPath.HostIds = accessPath.getIntegerListValue(accessPath.HostIDs)
	}
	if accessPath.ProtectionDomainID != nil {
		req.AccessPath.ProtectionDomainID = accessPath.getIntegerValue(accessPath.ProtectionDomainID)
	}
	if accessPath.Tname != nil {
		req.AccessPath.Tname = accessPath.getStringValue(accessPath.Tname)
	}
	if accessPath.Tsecret != nil {
		req.AccessPath.Tsecret = accessPath.getStringValue(accessPath.Tsecret)
	}
	if accessPath.Type != nil {
		req.AccessPath.Type = accessPath.getStringValue(accessPath.Type)
	}

	for _, mappingGroup := range accessPath.MappingGroups {
		mappingGroupReq := new(mappingGroupReq)
		if mappingGroup.BlockVolumeIDs != nil {
			mappingGroupReq.BlockVolumeIds = accessPath.getIntegerListValue(mappingGroup.BlockVolumeIDs)
		}
		if mappingGroup.ClientGroupID != nil {
			mappingGroupReq.ClientGroupID = accessPath.getIntegerValue(mappingGroup.ClientGroupID)
		}

		req.AccessPath.MappingGroups = append(req.AccessPath.MappingGroups, *mappingGroupReq)
	}

	body, err := accessPath.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create access path %s", name)
	}
	id, status, err := accessPath.getIdentifyAndStatus(body)
	accessPath.repr = id
	return accessPath.checkStatus(status)
}
