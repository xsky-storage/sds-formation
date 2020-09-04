package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type poolRuleReq struct {
	PlacementNodeID int64 `json:"placement_node_id,omitempty"`

	ReplicateNum int64 `json:"replicate_num,omitempty"`
}

// PoolCreateReq defines pool create request
type PoolCreateReq struct {
	Pool struct {
		CodingChunkNum int64 `json:"coding_chunk_num,omitempty"`

		DataChunkNum int64 `json:"data_chunk_num,omitempty"`

		FailureDomainType string `json:"failure_domain_type,omitempty"`

		Name string `json:"name,omitempty"`

		PoolName string `json:"pool_name,omitempty"`

		OsdIds []int64 `json:"osd_ids,omitempty"`

		PoolRole string `json:"pool_role,omitempty"`

		PoolType string `json:"pool_type,omitempty"`

		PrimaryPlacementNodeID int64 `json:"primary_placement_node_id,omitempty"`

		ProtectionDomainID int64 `json:"protection_domain_id,omitempty"`

		Ruleset []poolRuleReq `json:"ruleset,omitempty"`

		Size int64 `json:"size,omitempty"`
	} `json:"pool"`
}

// Pool resource
type Pool struct {
	ResourceBase
	CodingChunkNum     *parser.IntegerExpr
	DataChunkNum       *parser.IntegerExpr
	FailureDomainType  *parser.StringExpr
	Name               *parser.StringExpr
	PoolName           *parser.StringExpr
	OsdIDs             *parser.IntegerListExpr
	PoolType           *parser.StringExpr
	PoolRole           *parser.StringExpr
	ProtectionDomainID *parser.IntegerExpr
	Size               *parser.IntegerExpr
}

// Init inits resource instance
func (pool *Pool) Init(stack utils.StackInterface) {
	pool.ResourceBase.Init(stack)
	pool.setDelegate(pool)
}

// GetType return resource type
func (pool *Pool) GetType() string {
	return utils.ResourcePool
}

// CheckInterval get check interval
func (pool *Pool) CheckInterval() int {
	return 10
}

// IsReady check if the formation args are ready
func (pool *Pool) IsReady() (ready bool) {
	if !pool.isReady(pool.CodingChunkNum) ||
		!pool.isReady(pool.DataChunkNum) ||
		!pool.isReady(pool.FailureDomainType) ||
		!pool.isReady(pool.Name) ||
		!pool.isReady(pool.PoolName) ||
		!pool.isReady(pool.OsdIDs) ||
		!pool.isReady(pool.PoolType) ||
		!pool.isReady(pool.ProtectionDomainID) ||
		!pool.isReady(pool.Size) {
		return false
	}

	return true
}

func (pool *Pool) fakeCreate() (bool, error) {
	pool.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (pool *Pool) Create() (created bool, err error) {
	if pool.Name == nil {
		err = errors.Errorf("Name is required for resource %s", pool.GetType())
		return
	}
	if config.DryRun {
		return pool.fakeCreate()
	}

	name := pool.getStringValue(pool.Name)
	resourceID, err := pool.getResourceByName(name)
	if err != nil {
		return false, errors.Annotatef(err, "get pool %s", name)
	}
	if resourceID != nil {
		pool.repr = resourceID
		return false, nil
	}

	req := new(PoolCreateReq)
	poolInfo := &req.Pool
	poolInfo.Name = name
	if pool.PoolName != nil {
		poolInfo.PoolName = pool.getStringValue(pool.PoolName)
	}
	if pool.CodingChunkNum != nil {
		poolInfo.CodingChunkNum = pool.getIntegerValue(pool.CodingChunkNum)
	}
	if pool.DataChunkNum != nil {
		poolInfo.DataChunkNum = pool.getIntegerValue(pool.DataChunkNum)
	}
	if pool.FailureDomainType != nil {
		poolInfo.FailureDomainType = pool.getStringValue(pool.FailureDomainType)
	}
	if pool.OsdIDs != nil {
		poolInfo.OsdIds = pool.getIntegerListValue(pool.OsdIDs)
	}
	if pool.PoolType != nil {
		poolInfo.PoolType = pool.getStringValue(pool.PoolType)
	}
	if pool.PoolRole != nil {
		poolInfo.PoolRole = pool.getStringValue(pool.PoolRole)
	} else {
		poolInfo.PoolRole = utils.PoolOsdRoleData
	}
	if pool.ProtectionDomainID != nil {
		poolInfo.ProtectionDomainID = pool.getIntegerValue(pool.ProtectionDomainID)
	}
	if pool.Size != nil {
		poolInfo.Size = pool.getIntegerValue(pool.Size)
	}

	body, err := pool.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create pool %s", name)
	}

	id, _, err := pool.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	pool.repr = id

	return false, nil
}
