package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSUserCreateReq defines fs user create request
type FSUserCreateReq struct {
	FSUser struct {
		// type of file storage user
		Type string `json:"type,omitempty"`
		// name of file storage user
		Name string `json:"name,omitempty"`
		// email of file storage user
		Email string `json:"email,omitempty"`
		// password of file storage user
		Password string `json:"password,omitempty"`
		// id of file storage ad user
		FSADUserID int64 `json:"fs_ad_user_id,omitempty"`
		// id of file storage ldap user
		FSLdapUserID int64 `json:"fs_ldap_user_id,omitempty"`
	} `json:"fs_user"`
}

// FSUser resource
type FSUser struct {
	ResourceBase

	Type         *parser.StringExpr
	Name         *parser.StringExpr
	Email        *parser.StringExpr
	Password     *parser.StringExpr
	FSADUserID   *parser.IntegerExpr
	FSLdapUserID *parser.IntegerExpr
}

// Init inits resource instance
func (user *FSUser) Init(stack utils.StackInterface) {
	user.ResourceBase.Init(stack)
	user.setDelegate(user)
}

// GetType return resource type
func (user *FSUser) GetType() string {
	return utils.ResourceFSUser
}

// IsReady check if the formation args are ready
func (user *FSUser) IsReady() (ready bool) {
	if !user.isReady(user.Type) ||
		!user.isReady(user.Name) ||
		!user.isReady(user.Email) ||
		!user.isReady(user.Password) ||
		!user.isReady(user.FSADUserID) ||
		!user.isReady(user.FSLdapUserID) {
		return false
	}

	return true
}

func (user *FSUser) fakeCreate() (bool, error) {
	user.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (user *FSUser) Create() (created bool, err error) {
	if user.Name == nil {
		return false, errors.Errorf("Name is required for resource %s", user.GetType())
	}
	if config.DryRun {
		return user.fakeCreate()
	}

	name := user.getStringValue(user.Name)
	resourceID, err := user.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		user.repr = resourceID
		return true, nil
	}

	req := new(FSUserCreateReq)
	userInfo := &req.FSUser
	if user.Type != nil {
		userInfo.Type = user.getStringValue(user.Type)
	}
	if user.Name != nil {
		userInfo.Name = user.getStringValue(user.Name)
	}
	if user.Email != nil {
		userInfo.Email = user.getStringValue(user.Email)
	}
	if user.Password != nil {
		userInfo.Password = user.getStringValue(user.Password)
	}
	if user.FSADUserID != nil {
		userInfo.FSADUserID = user.getIntegerValue(user.FSADUserID)
	}
	if user.FSLdapUserID != nil {
		userInfo.FSLdapUserID = user.getIntegerValue(user.FSLdapUserID)
	}

	body, err := user.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs user %s", name)
	}

	id, _, err := user.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	user.repr = id
	return true, nil
}
