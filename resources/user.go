package formation

import (
	"fmt"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// UserCreateReq defines user create request
type UserCreateReq struct {
	User struct {
		// email of user
		Email string `json:"email"`
		// enable or disable the user
		Enabled bool `json:"enabled"`
		// name of user
		Name string `json:"name"`
		// password of user
		Password string `json:"password"`
		// roles of user
		Roles []string `json:"roles,omitempty"`
	} `json:"user"`
}

// User resource
type User struct {
	ResourceBase
	Name     *parser.StringExpr
	Email    *parser.StringExpr
	Password *parser.StringExpr
	Enabled  *parser.BoolExpr
}

// Init inits resource instance
func (user *User) Init(stack utils.StackInterface) {
	user.ResourceBase.Init(stack)
	user.setDelegate(user)
}

// GetType return resource type
func (user *User) GetType() string {
	return utils.ResourceUser
}

// IsReady check if the formation args are ready
func (user *User) IsReady() (ready bool) {
	if !user.isReady(user.Name) ||
		!user.isReady(user.Email) ||
		!user.isReady(user.Password) ||
		!user.isReady(user.Enabled) {
		return false
	}

	return true
}

func (user *User) fakeCreate() (bool, error) {
	user.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (user *User) Create() (created bool, err error) {
	if user.Name == nil {
		err = fmt.Errorf("Name is required for resource %s", user.GetType())
		return
	}
	if config.DryRun {
		return user.fakeCreate()
	}

	name := user.getStringValue(user.Name)
	resourceID, err := user.getResourceByName(name)
	if err != nil {
		return false, errors.Annotatef(err, "get user %s", name)
	}
	if resourceID != nil {
		user.repr = resourceID
		return true, nil
	}

	req := new(UserCreateReq)
	userInfo := &req.User
	userInfo.Name = name
	if user.Email != nil {
		userInfo.Email = user.getStringValue(user.Email)
	}
	if user.Password != nil {
		userInfo.Password = user.getStringValue(user.Password)
	}
	if user.Enabled != nil {
		userInfo.Enabled = user.getBoolValue(user.Enabled)
	}

	body, err := user.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create user %s", name)
	}
	id, _, err := user.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Annotate(err, "parse server data")
	}
	user.repr = id

	return true, nil
}
