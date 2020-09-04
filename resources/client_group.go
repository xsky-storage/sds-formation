package formation

import (
	"fmt"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type clientGroupCreateReqClientGroupClientsElt struct {
	Code string `json:"code,omitempty"`
}

// ClientGroupCreateReq defines Client group create request
type ClientGroupCreateReq struct {
	ClientGroup struct {
		Chap bool `json:"chap,omitempty"`

		Clients []clientGroupCreateReqClientGroupClientsElt `json:"clients,omitempty"`

		Description string `json:"description,omitempty"`

		Iname string `json:"iname,omitempty"`

		Isecret string `json:"isecret,omitempty"`

		Name string `json:"name,omitempty"`

		Type string `json:"type,omitempty"`
	} `json:"client_group"`
}

// Client resource
type Client struct {
	Code *parser.StringExpr
}

// ClientGroup resource
type ClientGroup struct {
	ResourceBase

	Clients     []*Client
	Description *parser.StringExpr
	Name        *parser.StringExpr
	Type        *parser.StringExpr
}

// Init inits resource instance
func (clientGroup *ClientGroup) Init(stack utils.StackInterface) {
	clientGroup.ResourceBase.Init(stack)
	clientGroup.setDelegate(clientGroup)
}

// GetType return resource type
func (clientGroup *ClientGroup) GetType() string {
	return utils.ResourceClientGroup
}

// IsReady check if the formation args are ready
func (clientGroup *ClientGroup) IsReady() (ready bool) {
	if !clientGroup.isReady(clientGroup.Description) ||
		!clientGroup.isReady(clientGroup.Name) ||
		!clientGroup.isReady(clientGroup.Type) {
		return false
	}
	for _, client := range clientGroup.Clients {
		if !clientGroup.isReady(client.Code) {
			return false
		}
	}

	return true
}

func (clientGroup *ClientGroup) fakeCreate() (bool, error) {
	clientGroup.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (clientGroup *ClientGroup) Create() (created bool, err error) {
	if clientGroup.Name == nil {
		err = fmt.Errorf("Name is required for resource %s", clientGroup.GetType())
		return
	}
	if config.DryRun {
		return clientGroup.fakeCreate()
	}

	name := clientGroup.getStringValue(clientGroup.Name)
	resourceID, err := clientGroup.getResourceByName(name)
	if err != nil {
		return false, errors.Annotatef(err, "check if client group %s exists", name)
	}
	if resourceID != nil {
		clientGroup.repr = resourceID
		return false, nil
	}

	req := new(ClientGroupCreateReq)
	if clientGroup.Description != nil {
		req.ClientGroup.Description = clientGroup.getStringValue(clientGroup.Description)
	}
	if clientGroup.Name != nil {
		req.ClientGroup.Name = clientGroup.getStringValue(clientGroup.Name)
	}
	if clientGroup.Type != nil {
		req.ClientGroup.Type = clientGroup.getStringValue(clientGroup.Type)
	}

	for _, c := range clientGroup.Clients {
		clientReq := new(clientGroupCreateReqClientGroupClientsElt)
		if c.Code != nil {
			clientReq.Code = clientGroup.getStringValue(c.Code)
		}

		req.ClientGroup.Clients = append(req.ClientGroup.Clients, *clientReq)
	}

	body, err := clientGroup.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create client group %s", name)
	}
	id, status, err := clientGroup.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	clientGroup.repr = id
	return clientGroup.checkStatus(status)
}
