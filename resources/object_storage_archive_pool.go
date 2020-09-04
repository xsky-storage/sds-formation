package formation

import (
	"encoding/json"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// OSArchivePoolCreateReq defines os archive pool create request
type OSArchivePoolCreateReq struct {
	ArchivePool struct {
		PoolID int64 `json:"pool_id"`
	} `json:"os_archive_pool"`
}

// ObjectStorageArchivePool resource
type ObjectStorageArchivePool struct {
	ResourceBase

	PoolID *parser.IntegerExpr
}

// Init inits resource instance
func (pool *ObjectStorageArchivePool) Init(stack utils.StackInterface) {
	pool.ResourceBase.Init(stack)
	pool.setDelegate(pool)
}

// GetType return resource type
func (pool *ObjectStorageArchivePool) GetType() string {
	return utils.ResourceObjectStorageArchivePool
}

// IsReady check if the formation args are ready
func (pool *ObjectStorageArchivePool) IsReady() (ready bool) {
	if !pool.isReady(pool.PoolID) {
		return false
	}

	return true
}

func (pool *ObjectStorageArchivePool) fakeCreate() (bool, error) {
	pool.repr = rand.Int63()
	return true, nil
}

func (pool *ObjectStorageArchivePool) getResource(poolID int64) (resourceID int64, err error) {
	body, err := pool.CallResourceAPI(utils.ListAPIName, nil, nil)
	if err != nil {
		return 0, errors.Annotatef(err, "list archive pools")
	}

	records := &struct {
		ArchivePools []struct {
			ID   int64 `json:"id"`
			Pool struct {
				ID int64 `json:"id"`
			} `json:"pool"`
		} `json:"os_archive_pools"`
	}{}
	if err = json.Unmarshal(body, records); err != nil {
		return 0, errors.Trace(err)
	}
	for _, archivePoolResp := range records.ArchivePools {
		if archivePoolResp.Pool.ID == poolID {
			return archivePoolResp.ID, nil
		}
	}
	return 0, nil
}

// Create create the resource
func (pool *ObjectStorageArchivePool) Create() (created bool, err error) {
	if pool.PoolID == nil {
		return false, errors.Errorf("PoolID is required for resource %s", pool.GetType())
	}
	if config.DryRun {
		return pool.fakeCreate()
	}

	poolID := pool.getIntegerValue(pool.PoolID)
	resourceID, err := pool.getResource(poolID)
	if err != nil {
		return false, errors.Annotatef(err, "failed to get pool with id %d", poolID)
	}
	if resourceID > 0 {
		pool.repr = resourceID
		return true, nil
	}

	req := new(OSArchivePoolCreateReq)
	req.ArchivePool.PoolID = pool.getIntegerValue(pool.PoolID)
	body, err := pool.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err,
			"create object storage archive pool using pool with id %d", poolID)
	}

	id, _, err := pool.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	pool.repr = id
	return false, nil
}
