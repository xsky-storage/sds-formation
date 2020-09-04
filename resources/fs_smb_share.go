package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSSMBShareACLReq defines fs smb share acl creation or updation struct
type FSSMBShareACLReq struct {
	// id of share group
	ID int64 `json:"id,omitempty"`
	// type of share acl
	Type string `json:"type"`
	// id of user
	UserID int64 `json:"fs_user_id,omitempty"`
	// id of share group
	UserGroupID   int64  `json:"fs_user_group_id,omitempty"`
	UserName      string `json:"user_name,omitempty"`
	UserGroupName string `json:"user_group_name,omitempty"`
	// readonly or readwrite access
	Permission string `json:"permission"`
}

// FSSMBShareCreateReq defines request for fs smb share creation
type FSSMBShareCreateReq struct {
	Share struct {
		// name of share
		Name *string `json:"name"`
		// folder id
		FolderID int64 `json:"fs_folder_id" required:"true"`
		// quota tree id
		QuotaTreeID *int64 `json:"fs_quota_tree_id,omitempty"`
		// gateway group id
		GatewayGroupID int64 `json:"fs_gateway_group_id" required:"true"`
		// recycle status
		Recycled *bool `json:"recycled,omitempty"`
		// default acl status
		ACLInherited *bool `json:"acl_inherited,omitempty"`
		// case sensitive
		CaseSensitive *bool `json:"case_sensitive,omitempty"`
		// access control array
		ACLs []*FSSMBShareACLReq `json:"fs_smb_share_acls"`
	} `json:"fs_smb_share" required:"true"`
}

type fsSMBSahreACL struct {
	ID            *parser.IntegerExpr
	Type          *parser.StringExpr
	UserID        *parser.IntegerExpr
	UserGroupID   *parser.IntegerExpr
	UserName      *parser.StringExpr
	UserGroupName *parser.StringExpr
	Permission    *parser.StringExpr
}

// FSSMBShare resource
type FSSMBShare struct {
	ResourceBase

	Name           *parser.StringExpr
	FolderID       *parser.IntegerExpr
	QuotaTreeID    *parser.IntegerExpr
	GatewayGroupID *parser.IntegerExpr
	Recycled       *parser.BoolExpr
	ACLInherited   *parser.BoolExpr
	CaseSensitive  *parser.BoolExpr
	ACLs           []*fsSMBSahreACL
}

// Init inits resource instance
func (share *FSSMBShare) Init(stack utils.StackInterface) {
	share.ResourceBase.Init(stack)
	share.setDelegate(share)
}

// GetType return resource type
func (share *FSSMBShare) GetType() string {
	return utils.ResourceFSSMBShare
}

// IsReady check if the formation args are ready
func (share *FSSMBShare) IsReady() (ready bool) {
	if !share.isReady(share.Name) ||
		!share.isReady(share.FolderID) ||
		!share.isReady(share.QuotaTreeID) ||
		!share.isReady(share.GatewayGroupID) ||
		!share.isReady(share.Recycled) ||
		!share.isReady(share.ACLInherited) ||
		!share.isReady(share.CaseSensitive) {
		return false
	}
	for _, acl := range share.ACLs {
		if !share.isReady(acl.ID) ||
			!share.isReady(acl.Type) ||
			!share.isReady(acl.UserID) ||
			!share.isReady(acl.UserName) ||
			!share.isReady(acl.UserGroupName) ||
			!share.isReady(acl.UserGroupID) ||
			!share.isReady(acl.Permission) {
			return false
		}
	}

	return true
}

func (share *FSSMBShare) fakeCreate() (bool, error) {
	share.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (share *FSSMBShare) Create() (created bool, err error) {
	if share.Name == nil {
		return false, errors.Errorf("Name is required for resource %s", share.GetType())
	}
	if config.DryRun {
		return share.fakeCreate()
	}

	name := share.getStringValue(share.Name)
	resourceID, err := share.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		share.repr = resourceID
		return false, nil
	}

	req := new(FSSMBShareCreateReq)
	shareInfo := &req.Share
	shareInfo.Name = &name
	if share.FolderID != nil {
		shareInfo.FolderID = share.getIntegerValue(share.FolderID)
	}
	if share.QuotaTreeID != nil {
		quotaTreeID := share.getIntegerValue(share.QuotaTreeID)
		shareInfo.QuotaTreeID = &quotaTreeID
	}
	if share.GatewayGroupID != nil {
		shareInfo.GatewayGroupID = share.getIntegerValue(share.GatewayGroupID)
	}
	if share.Recycled != nil {
		recycled := share.getBoolValue(share.Recycled)
		shareInfo.Recycled = &recycled
	}
	if share.ACLInherited != nil {
		inherited := share.getBoolValue(share.ACLInherited)
		shareInfo.ACLInherited = &inherited
	}
	if share.CaseSensitive != nil {
		caseSensitive := share.getBoolValue(share.CaseSensitive)
		shareInfo.CaseSensitive = &caseSensitive
	}
	for _, acl := range share.ACLs {
		aclReq := new(FSSMBShareACLReq)
		if acl.ID != nil {
			aclReq.ID = share.getIntegerValue(acl.ID)
		}
		if acl.Type != nil {
			aclReq.Type = share.getStringValue(acl.Type)
		}
		if acl.UserID != nil {
			aclReq.UserID = share.getIntegerValue(acl.UserID)
		}
		if acl.UserGroupID != nil {
			aclReq.UserGroupID = share.getIntegerValue(acl.UserGroupID)
		}
		if acl.Permission != nil {
			aclReq.Permission = share.getStringValue(acl.Permission)
		}
		if acl.UserName != nil {
			aclReq.UserName = share.getStringValue(acl.UserName)
		}
		if acl.UserGroupName != nil {
			aclReq.UserGroupName = share.getStringValue(acl.UserGroupName)
		}
		shareInfo.ACLs = append(shareInfo.ACLs, aclReq)
	}

	body, err := share.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs smb share %s", name)
	}

	id, status, err := share.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	share.repr = id
	created, err = share.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
