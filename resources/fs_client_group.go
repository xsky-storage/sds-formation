package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSClientGroupCreateReq define request for file storage client group creation
type FSClientGroupCreateReq struct {
	ClientGroup struct {
		// name of client group
		Name string `json:"name" required:"true"`
		// ids of clients
		ClientIDs []int64 `json:"fs_client_ids" required:"true"`
	} `json:"fs_client_group" required:"true"`
}

// FSClientGroup resource
type FSClientGroup struct {
	ResourceBase

	Name      *parser.StringExpr
	ClientIDs *parser.IntegerListExpr
}

// Init inits resource instance
func (group *FSClientGroup) Init(stack utils.StackInterface) {
	group.ResourceBase.Init(stack)
	group.setDelegate(group)
}

// GetType return resource type
func (group *FSClientGroup) GetType() string {
	return utils.ResourceFSClientGroup
}

// IsReady check if the formation args are ready
func (group *FSClientGroup) IsReady() (ready bool) {
	if !group.isReady(group.Name) ||
		!group.isReady(group.ClientIDs) {
		return false
	}
	return true
}

func (group *FSClientGroup) fakeCreate() (bool, error) {
	group.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (group *FSClientGroup) Create() (created bool, err error) {
	if group.Name == nil {
		err = errors.Errorf("Name is required for resource %s", group.GetType())
		return
	}
	if config.DryRun {
		return group.fakeCreate()
	}

	name := group.getStringValue(group.Name)
	resourceID, err := group.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		group.repr = resourceID
		return true, nil
	}

	req := new(FSClientGroupCreateReq)
	groupInfo := &req.ClientGroup
	if group.Name != nil {
		groupInfo.Name = group.getStringValue(group.Name)
	}
	if group.ClientIDs != nil {
		groupInfo.ClientIDs = group.getIntegerListValue(group.ClientIDs)
	}

	body, err := group.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs client group %s", name)
	}

	id, _, err := group.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	group.repr = id
	return true, nil
}
