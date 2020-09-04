package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// OSCreateReq defines os initialize request
type OSCreateReq struct {
	ObjectStorage struct {

		// object storage archive pool id
		ArchivePoolID int64 `json:"archive_pool_id"`

		// system pool id
		PoolID int64 `json:"pool_id"`
	} `json:"object_storage"`
}

// ObjectStorage resource
type ObjectStorage struct {
	ResourceBase
	PoolID        *parser.IntegerExpr
	ArchivePoolID *parser.IntegerExpr
}

// Init inits resource instance
func (os *ObjectStorage) Init(stack utils.StackInterface) {
	os.ResourceBase.Init(stack)
	os.setDelegate(os)
}

// GetType return resource type
func (os *ObjectStorage) GetType() string {
	return utils.ResourceObjectStorage
}

// CheckInterval return CheckInterval
func (os *ObjectStorage) CheckInterval() int {
	return 6
}

// IsReady check if the formation args are ready
func (os *ObjectStorage) IsReady() (ready bool) {
	if !os.isReady(os.PoolID) ||
		!os.isReady(os.ArchivePoolID) {
		return false
	}

	return true
}

func (os *ObjectStorage) fakeCreate() (bool, error) {
	os.repr = rand.Int63()
	return true, nil
}

func (os *ObjectStorage) getResource() (resourceID int64, err error) {
	body, err := os.CallGetAPI()
	if err != nil {
		return 0, errors.Annotatef(err, "get object storage")
	}

	id, status, err := os.getIdentifyAndStatus(body)
	if err != nil {
		return 0, errors.Trace(err)
	}

	if status != utils.StatusUninitialized {
		return id.(int64), nil
	}
	return 0, nil
}

// Create create the resource
func (os *ObjectStorage) Create() (created bool, err error) {
	if config.DryRun {
		return os.fakeCreate()
	}

	resourceID, err := os.getResource()
	if err != nil {
		err = errors.Annotatef(err, "failed to get object storage")
		return
	}
	if resourceID > 0 {
		os.repr = resourceID
		return false, nil
	}

	req := new(OSCreateReq)
	osInfo := &req.ObjectStorage
	if os.PoolID != nil {
		osInfo.PoolID = os.getIntegerValue(os.PoolID)
	}
	if os.ArchivePoolID != nil {
		osInfo.ArchivePoolID = os.getIntegerValue(os.ArchivePoolID)
	}
	body, err := os.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "init object storage")
	}

	id, _, err := os.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	os.repr = id

	return false, nil
}
