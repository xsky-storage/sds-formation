package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSNFSShareACLReq defines fs nfs shares acl creation or updation struct
type FSNFSShareACLReq struct {
	// id of nfsShare group
	ID int64 `json:"id,omitempty"`
	// type of share acl
	Type string `json:"type"`
	// id of cilent
	ClientID int64 `json:"fs_client_id,omitempty"`
	// id of cilent group
	ClientGroupID int64 `json:"fs_client_group_id,omitempty"`
	// readonly or readwrite access
	Permission string `json:"permission"`
	// write to disk by sync or async
	Sync *bool `json:"sync,omitempty"`
	// all squash
	AllSquash *bool `json:"all_squash,omitempty"`
	// root squash
	RootSquash *bool `json:"root_squash,omitempty"`
}

// FSNFSShareCreateReq defines request for fs nfs shares creation
type FSNFSShareCreateReq struct {
	Share struct {
		// folder id
		FolderID int64 `json:"fs_folder_id" required:"true"`
		// quota tree id
		QuotaTreeID *int64 `json:"fs_quota_tree_id,omitempty"`
		// gateway group id
		GatewayGroupID int64 `json:"fs_gateway_group_id" required:"true"`
		// access control array
		ACLs []*FSNFSShareACLReq `json:"fs_nfs_share_acls,omitempty"`
	} `json:"fs_nfs_share" required:"true"`
}

// FSNFSShareACL resource
type FSNFSShareACL struct {
	ID            *parser.IntegerExpr
	Type          *parser.StringExpr
	ClientID      *parser.IntegerExpr
	ClientGroupID *parser.IntegerExpr
	Permission    *parser.StringExpr
	Sync          *parser.BoolExpr
	AllSquash     *parser.BoolExpr
	RootSquash    *parser.BoolExpr
}

// FSNFSShare resource
type FSNFSShare struct {
	ResourceBase

	FolderID       *parser.IntegerExpr
	QuotaTreeID    *parser.IntegerExpr
	GatewayGroupID *parser.IntegerExpr
	ACLs           []*FSNFSShareACL
}

// Init inits resource instance
func (nfsShare *FSNFSShare) Init(stack utils.StackInterface) {
	nfsShare.ResourceBase.Init(stack)
	nfsShare.setDelegate(nfsShare)
}

// GetType return resource type
func (nfsShare *FSNFSShare) GetType() string {
	return utils.ResourceFSNFSShare
}

// IsReady check if the formation args are ready
func (nfsShare *FSNFSShare) IsReady() (ready bool) {
	if !nfsShare.isReady(nfsShare.FolderID) ||
		!nfsShare.isReady(nfsShare.QuotaTreeID) ||
		!nfsShare.isReady(nfsShare.GatewayGroupID) {
		return false
	}
	for _, acl := range nfsShare.ACLs {
		if !nfsShare.isReady(acl.ID) ||
			!nfsShare.isReady(acl.ClientID) ||
			!nfsShare.isReady(acl.ClientGroupID) ||
			!nfsShare.isReady(acl.Permission) ||
			!nfsShare.isReady(acl.Sync) ||
			!nfsShare.isReady(acl.AllSquash) ||
			!nfsShare.isReady(acl.RootSquash) ||
			!nfsShare.isReady(acl.Type) {
			return false
		}
	}

	return true
}

func (nfsShare *FSNFSShare) fakeCreate() (bool, error) {
	nfsShare.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (nfsShare *FSNFSShare) Create() (created bool, err error) {
	if config.DryRun {
		return nfsShare.fakeCreate()
	}

	// TODO: check get by unique idenfity

	req := new(FSNFSShareCreateReq)
	userInfo := &req.Share
	if nfsShare.FolderID != nil {
		userInfo.FolderID = nfsShare.getIntegerValue(nfsShare.FolderID)
	}
	if nfsShare.QuotaTreeID != nil {
		quotaTreeID := nfsShare.getIntegerValue(nfsShare.QuotaTreeID)
		userInfo.QuotaTreeID = &quotaTreeID
	}
	if nfsShare.GatewayGroupID != nil {
		userInfo.GatewayGroupID = nfsShare.getIntegerValue(nfsShare.GatewayGroupID)
	}
	for _, acl := range nfsShare.ACLs {
		aclReq := new(FSNFSShareACLReq)

		if acl.ID != nil {
			aclReq.ID = nfsShare.getIntegerValue(acl.ID)
		}
		if acl.Type != nil {
			aclReq.Type = nfsShare.getStringValue(acl.Type)
		}
		if acl.ClientID != nil {
			aclReq.ClientID = nfsShare.getIntegerValue(acl.ClientID)
		}
		if acl.ClientGroupID != nil {
			aclReq.ClientGroupID = nfsShare.getIntegerValue(acl.ClientGroupID)
		}

		if acl.Permission != nil {
			aclReq.Permission = nfsShare.getStringValue(acl.Permission)
		}

		if acl.Sync != nil {
			sync := nfsShare.getBoolValue(acl.Sync)
			aclReq.Sync = &sync
		}
		if acl.AllSquash != nil {
			squash := nfsShare.getBoolValue(acl.AllSquash)
			aclReq.AllSquash = &squash
		}
		if acl.RootSquash != nil {
			rootSquash := nfsShare.getBoolValue(acl.RootSquash)
			aclReq.RootSquash = &rootSquash
		}
		userInfo.ACLs = append(userInfo.ACLs, aclReq)
	}

	body, err := nfsShare.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs nfs share")
	}

	id, status, err := nfsShare.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	nfsShare.repr = id
	created, err = nfsShare.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
