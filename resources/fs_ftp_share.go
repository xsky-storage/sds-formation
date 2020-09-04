package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSFTPShareACLReq defines fs ftp share acl creation or updation struct
type FSFTPShareACLReq struct {
	// id of share group
	ID int64 `json:"id,omitempty"`
	// type of share acl
	Type string `json:"type"`
	// id of user
	UserID int64 `json:"fs_user_id,omitempty"`
	// id of share group
	UserGroupID int64 `json:"fs_user_group_id,omitempty"`
	// enable listing files
	ListEnabled bool `json:"list_enabled"`
	// enable creating files
	CreateEnabled bool `json:"create_enabled"`
	// enable renaming files
	RenameEnabled bool `json:"rename_enabled"`
	// enable deleting files
	DeleteEnabled bool `json:"delete_enabled"`
	// enable uploading files
	UploadEnabled bool `json:"upload_enabled"`
	// max bandwidth of uploading
	UploadBandwidth uint64 `json:"upload_bandwidth,omitempty"`
	// enable downloading files
	DownloadEnabled bool `json:"download_enabled"`
	// max bandwidth of downloading
	DownloadBandwidth uint64 `json:"download_bandwidth,omitempty"`
}

// FSFTPShareCreateReq defines request for fs ftp share creation
type FSFTPShareCreateReq struct {
	Share struct {
		// name of share
		Name *string `json:"name"`
		// folder id
		FolderID int64 `json:"fs_folder_id"`
		// quota tree id
		QuotaTreeID *int64 `json:"fs_quota_tree_id,omitempty"`
		// gateway group id
		GatewayGroupID int64 `json:"fs_gateway_group_id"`
		// access control array
		ACLs []*FSFTPShareACLReq `json:"fs_ftp_share_acls,omitempty"`
	} `json:"fs_ftp_share" required:"true"`
}

// FSFTPShareACL resource acl
type FSFTPShareACL struct {
	ResourceBase

	ID                *parser.IntegerExpr
	Type              *parser.StringExpr
	UserID            *parser.IntegerExpr
	UserGroupID       *parser.IntegerExpr
	ListEnabled       *parser.BoolExpr
	CreateEnabled     *parser.BoolExpr
	RenameEnabled     *parser.BoolExpr
	DeleteEnabled     *parser.BoolExpr
	UploadEnabled     *parser.BoolExpr
	UploadBandwidth   *parser.IntegerExpr
	DownloadEnabled   *parser.BoolExpr
	DownloadBandwidth *parser.IntegerExpr
}

// FSFTPShare resource
type FSFTPShare struct {
	ResourceBase

	Name           *parser.StringExpr
	FolderID       *parser.IntegerExpr
	QuotaTreeID    *parser.IntegerExpr
	GatewayGroupID *parser.IntegerExpr
	ACLs           []*FSFTPShareACL
}

// Init inits resource instance
func (share *FSFTPShare) Init(stack utils.StackInterface) {
	share.ResourceBase.Init(stack)
	share.setDelegate(share)
}

// GetType return resource type
func (share *FSFTPShare) GetType() string {
	return utils.ResourceFSFTPShare
}

// IsReady check if the formation args are ready
func (share *FSFTPShare) IsReady() (ready bool) {
	if !share.isReady(share.Name) ||
		!share.isReady(share.FolderID) ||
		!share.isReady(share.QuotaTreeID) ||
		!share.isReady(share.GatewayGroupID) {
		return false
	}
	for _, acl := range share.ACLs {
		if !share.isReady(acl.ID) ||
			!share.isReady(acl.Type) ||
			!share.isReady(acl.UserGroupID) ||
			!share.isReady(acl.UserID) ||
			!share.isReady(acl.ListEnabled) ||
			!share.isReady(acl.CreateEnabled) ||
			!share.isReady(acl.RenameEnabled) ||
			!share.isReady(acl.DeleteEnabled) ||
			!share.isReady(acl.DownloadEnabled) ||
			!share.isReady(acl.DownloadBandwidth) ||
			!share.isReady(acl.UploadBandwidth) ||
			!share.isReady(acl.UploadEnabled) {
			return false
		}
	}

	return true
}

func (share *FSFTPShare) fakeCreate() (bool, error) {
	share.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (share *FSFTPShare) Create() (created bool, err error) {
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

	req := new(FSFTPShareCreateReq)
	ftpShare := &req.Share
	if share.Name != nil {
		name := share.getStringValue(share.Name)
		ftpShare.Name = &name
	}
	if share.FolderID != nil {
		ftpShare.FolderID = share.getIntegerValue(share.FolderID)
	}
	if share.QuotaTreeID != nil {
		quotaTreeID := share.getIntegerValue(share.QuotaTreeID)
		ftpShare.QuotaTreeID = &quotaTreeID
	}
	if share.GatewayGroupID != nil {
		ftpShare.GatewayGroupID = share.getIntegerValue(share.GatewayGroupID)
	}
	for _, acl := range share.ACLs {
		aclReq := new(FSFTPShareACLReq)
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
		if acl.ListEnabled != nil {
			aclReq.ListEnabled = share.getBoolValue(acl.ListEnabled)
		}
		if acl.CreateEnabled != nil {
			aclReq.CreateEnabled = share.getBoolValue(acl.CreateEnabled)
		}
		if acl.RenameEnabled != nil {
			aclReq.RenameEnabled = share.getBoolValue(acl.RenameEnabled)
		}
		if acl.DeleteEnabled != nil {
			aclReq.DeleteEnabled = share.getBoolValue(acl.DeleteEnabled)
		}
		if acl.UploadEnabled != nil {
			aclReq.UploadEnabled = share.getBoolValue(acl.UploadEnabled)
		}
		if acl.UploadBandwidth != nil {
			aclReq.UploadBandwidth = uint64(share.getIntegerValue(acl.UploadBandwidth))
		}
		if acl.DownloadEnabled != nil {
			aclReq.DownloadEnabled = share.getBoolValue(acl.DownloadEnabled)
		}
		if acl.DownloadBandwidth != nil {
			aclReq.DownloadBandwidth = uint64(share.getIntegerValue(acl.DownloadBandwidth))
		}
		ftpShare.ACLs = append(ftpShare.ACLs, aclReq)
	}

	body, err := share.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs ftp share %s", name)
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
