package formation

import (
	"encoding/json"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// OsdCreateReq defines osd create request
type OsdCreateReq struct {
	Osd struct {

		// data disk id
		DiskID int64 `json:"disk_id,omitempty"`

		// cache partition id
		PartitionID int64 `json:"partition_id,omitempty"`

		OmapByte int64 `json:"omap_byte,omitempty"`

		// read cache size in bytes
		ReadCacheSize int64 `json:"read_cache_size,omitempty"`

		// osd role: \"data\" or \"index\", default is \"data\"
		Role string `json:"role,omitempty"`
	} `json:"osd"`
}

// Osd resource
type Osd struct {
	ResourceBase
	OmapByte    *parser.IntegerExpr
	DiskID      *parser.IntegerExpr
	PartitionID *parser.IntegerExpr
	Role        *parser.StringExpr
}

// Init inits resource instance
func (osd *Osd) Init(stack utils.StackInterface) {
	osd.ResourceBase.Init(stack)
	osd.setDelegate(osd)
}

// GetType return resource type
func (osd *Osd) GetType() string {
	return utils.ResourceOsd
}

// CheckInterval get check interval
func (osd *Osd) CheckInterval() int {
	return 10
}

// IsReady check if the formation args are ready
func (osd *Osd) IsReady() (ready bool) {
	if !osd.isReady(osd.DiskID) ||
		!osd.isReady(osd.PartitionID) {
		return false
	}

	return true
}

func (osd *Osd) fakeCreate() (bool, error) {
	osd.repr = rand.Int63()
	return true, nil
}

func (osd *Osd) getResource(diskID int64) (resourceID int64, err error) {
	body, err := osd.CallResourceAPI(utils.ListAPIName, nil, nil)
	if err != nil {
		return 0, errors.Annotatef(err, "list osds")
	}

	records := &struct {
		Osds []*struct {
			ID   int64 `json:"id"`
			Disk struct {
				ID int64 `json:"id"`
			} `json:"disk"`
		} `json:"osds"`
	}{}
	if err = json.Unmarshal(body, records); err != nil {
		return 0, errors.Annotate(err, "parse list osds data")
	}

	for _, osdResp := range records.Osds {
		if osdResp.Disk.ID == diskID {
			return osdResp.ID, nil
		}
	}
	return 0, nil
}

// Create create the resource
func (osd *Osd) Create() (created bool, err error) {
	if osd.DiskID == nil {
		return false, errors.Errorf("DiskID is required for resource %s", osd.GetType())
	}
	if config.DryRun {
		return osd.fakeCreate()
	}

	diskID := osd.getIntegerValue(osd.DiskID)
	resourceID, err := osd.getResource(diskID)
	if err != nil {
		return false, errors.Annotatef(err, "get osd with disk %d", diskID)
	}
	if resourceID > 0 {
		osd.repr = resourceID
		return false, nil
	}

	req := new(OsdCreateReq)
	osdInfo := &req.Osd
	osdInfo.DiskID = diskID
	if osd.PartitionID != nil {
		osdInfo.PartitionID = osd.getIntegerValue(osd.PartitionID)
	}
	if osd.Role != nil {
		osdInfo.Role = osd.getStringValue(osd.Role)
	} else {
		osdInfo.Role = utils.PoolOsdRoleData
	}
	if osd.OmapByte != nil {
		osdInfo.OmapByte = osd.getIntegerValue(osd.OmapByte)
	}

	body, err := osd.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create osd with disk %d", diskID)
	}

	id, _, err := osd.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	osd.repr = id

	return false, nil
}
