package formation

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// Token resource
type Token struct {
	ResourceBase
	Name     *parser.StringExpr
	Email    *parser.StringExpr
	Password *parser.StringExpr
}

// Init inits resource instance
func (token *Token) Init(stack utils.StackInterface) {
	token.ResourceBase.Init(stack)
	token.setDelegate(token)
}

// GetType return resource type
func (token *Token) GetType() string {
	return utils.ResourceToken
}

// IsReady check if the formation args are ready
func (token *Token) IsReady() (ready bool) {
	if !token.isReady(token.Name) ||
		!token.isReady(token.Email) ||
		!token.isReady(token.Password) {
		return false
	}

	return true
}

func (token *Token) fakeCreate() (bool, error) {
	token.repr = "28171a13317c4252806979bc86a69c4f"
	return true, nil
}

type authReq struct {
	User struct {
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		Password *string `json:"password,omitempty"`
	} `json:"user"`
}

// Create create the resource
func (token *Token) Create() (created bool, err error) {
	if config.DryRun {
		return token.fakeCreate()
	}

	req := new(authReq)
	if token.Name != nil {
		name := token.getStringValue(token.Name)
		req.User.Name = &name
	}
	if token.Email != nil {
		email := token.getStringValue(token.Email)
		req.User.Email = &email
	}
	if token.Password != nil {
		passwd := token.getStringValue(token.Password)
		req.User.Password = &passwd
	}
	data := new(struct {
		Auth struct {
			Identity struct {
				Password *authReq `json:"password"`
			} `json:"identity"`
		} `json:"auth"`
	})
	data.Auth.Identity.Password = req
	bytes, err := token.CallCreateAPI(data, nil)
	if err != nil {
		return false, errors.Trace(err)
	}
	resp := new(struct {
		Token struct {
			UUID string `json:"uuid"`
		} `json:"token"`
	})
	if err = json.Unmarshal(bytes, resp); err != nil {
		return false, errors.Trace(err)
	}
	token.repr = resp.Token.UUID

	return true, nil
}
