package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSUserGroupCreateReq defines os user create request
type FSUserGroupCreateReq struct {
	UserGroup struct {
		// name of user group
		Name string `json:"name"`
		// fs security type
		Type string `json:"type,omitempty"`
		// ids of users, which are required when type is smb or ftp
		UserIDs []int64 `json:"fs_user_ids"`
		// id of file storage ad user group
		FSADUserGroupID int64 `json:"fs_ad_user_group_id,omitempty"`
		// id of file storage ldap user group
		FSLdapUserGroupID int64 `json:"fs_ldap_user_group_id,omitempty"`
	} `json:"fs_user_group"`
}

// FSUserGroup resource
type FSUserGroup struct {
	ResourceBase

	Name              *parser.StringExpr
	Type              *parser.StringExpr
	UserIDs           *parser.IntegerListExpr
	FSADUserGroupID   *parser.IntegerExpr
	FSLdapUserGroupID *parser.IntegerExpr
}

// Init inits resource instance
func (userGroup *FSUserGroup) Init(stack utils.StackInterface) {
	userGroup.ResourceBase.Init(stack)
	userGroup.setDelegate(userGroup)
}

// GetType return resource type
func (userGroup *FSUserGroup) GetType() string {
	return utils.ResourceFSUserGroup
}

// IsReady check if the formation args are ready
func (userGroup *FSUserGroup) IsReady() (ready bool) {
	if !userGroup.isReady(userGroup.Name) ||
		!userGroup.isReady(userGroup.Type) ||
		!userGroup.isReady(userGroup.UserIDs) ||
		!userGroup.isReady(userGroup.FSADUserGroupID) ||
		!userGroup.isReady(userGroup.FSLdapUserGroupID) {
		return false
	}

	return true
}

func (userGroup *FSUserGroup) fakeCreate() (bool, error) {
	userGroup.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (userGroup *FSUserGroup) Create() (created bool, err error) {
	if userGroup.Name == nil {
		err = errors.Errorf("Name is required for resource %s", userGroup.GetType())
		return
	}
	if config.DryRun {
		return userGroup.fakeCreate()
	}

	name := userGroup.getStringValue(userGroup.Name)
	resourceID, err := userGroup.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		userGroup.repr = resourceID
		return true, nil
	}

	req := new(FSUserGroupCreateReq)
	userInfo := &req.UserGroup
	if userGroup.Name != nil {
		userInfo.Name = userGroup.getStringValue(userGroup.Name)
	}
	if userGroup.Type != nil {
		userInfo.Type = userGroup.getStringValue(userGroup.Type)
	}
	if userGroup.UserIDs != nil {
		userInfo.UserIDs = userGroup.getIntegerListValue(userGroup.UserIDs)
	}
	if userGroup.FSADUserGroupID != nil {
		userInfo.FSADUserGroupID = userGroup.getIntegerValue(userGroup.FSADUserGroupID)
	}
	if userGroup.FSLdapUserGroupID != nil {
		userInfo.FSLdapUserGroupID = userGroup.getIntegerValue(userGroup.FSLdapUserGroupID)
	}

	body, err := userGroup.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs user group %s", name)
	}

	id, _, err := userGroup.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	userGroup.repr = id
	return true, nil
}
