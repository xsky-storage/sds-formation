package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSArbitrationPoolCreateReq defines request to create file storage arbitration pool
type FSArbitrationPoolCreateReq struct {
	Info struct {
		// pool used for file system arbitration
		PoolID int64 `json:"pool_id" required:"true"`
	} `json:"fs_arbitration_pool" required:"true"`
}

// FSArbitrationPool resource
type FSArbitrationPool struct {
	ResourceBase

	PoolID *parser.IntegerExpr
}

// Init inits resource instance
func (abPool *FSArbitrationPool) Init(stack utils.StackInterface) {
	abPool.ResourceBase.Init(stack)
	abPool.setDelegate(abPool)
}

// GetType return resource type
func (abPool *FSArbitrationPool) GetType() string {
	return utils.ResourceFSArbitrationPool
}

// IsReady check if the formation args are ready
func (abPool *FSArbitrationPool) IsReady() (ready bool) {
	if !abPool.isReady(abPool.PoolID) {
		return false
	}

	return true
}

func (abPool *FSArbitrationPool) fakeCreate() (bool, error) {
	abPool.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (abPool *FSArbitrationPool) Create() (created bool, err error) {
	if config.DryRun {
		return abPool.fakeCreate()
	}

	req := new(FSArbitrationPoolCreateReq)
	if abPool.PoolID != nil {
		req.Info.PoolID = abPool.getIntegerValue(abPool.PoolID)
	}

	body, err := abPool.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs arbitration pool")
	}

	id, _, err := abPool.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	abPool.repr = id
	return true, nil
}
