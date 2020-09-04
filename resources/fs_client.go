package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSClientCreateReq define request struct for creating file storage client
type FSClientCreateReq struct {
	Client struct {
		// name of client
		Name string `json:"name" required:"true"`
		// ip of client
		IP string `json:"ip" required:"true"`
		// file storage client
	} `json:"fs_client" required:"true"`
}

// FSClient resource
type FSClient struct {
	ResourceBase

	Name *parser.StringExpr
	IP   *parser.StringExpr
}

// Init inits resource instance
func (client *FSClient) Init(stack utils.StackInterface) {
	client.ResourceBase.Init(stack)
	client.setDelegate(client)
}

// GetType return resource type
func (client *FSClient) GetType() string {
	return utils.ResourceFSClient
}

// IsReady check if the formation args are ready
func (client *FSClient) IsReady() (ready bool) {
	if !client.isReady(client.Name) ||
		!client.isReady(client.IP) {
		return false
	}

	return true
}

func (client *FSClient) fakeCreate() (bool, error) {
	client.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (client *FSClient) Create() (created bool, err error) {
	if client.Name == nil {
		err = errors.Errorf("Name is required for resource %s", client.GetType())
		return
	}
	if config.DryRun {
		return client.fakeCreate()
	}

	name := client.getStringValue(client.Name)
	resourceID, err := client.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		client.repr = resourceID
		return true, nil
	}

	req := new(FSClientCreateReq)
	clientInfo := &req.Client
	if client.Name != nil {
		clientInfo.Name = client.getStringValue(client.Name)
	}
	if client.IP != nil {
		clientInfo.IP = client.getStringValue(client.IP)
	}

	body, err := client.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs client %s", name)
	}

	id, status, err := client.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	client.repr = id
	created, err = client.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return true, nil
}
