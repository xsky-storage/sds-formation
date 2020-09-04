package formation

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// Osds resource
type Osds struct {
	ResourceBase
	DiskIDs        *parser.IntegerListExpr
	PartitionIDs   *parser.IntegerListExpr
	OmapByte       *parser.IntegerExpr
	Role           *parser.StringExpr
	creatingOsdIDs []int64
}

// Init inits resource instance
func (osds *Osds) Init(stack utils.StackInterface) {
	osds.ResourceBase.Init(stack)
	osds.setDelegate(osds)
}

// GetType return resource type
func (osds *Osds) GetType() string {
	return utils.ResourceOsds
}

// CheckInterval get check interval
func (osds *Osds) CheckInterval() int {
	return 10 * len(osds.creatingOsdIDs)
}

// IsReady check if the formation args are ready
func (osds *Osds) IsReady() (ready bool) {
	if !osds.isReady(osds.DiskIDs) {
		return false
	}

	return true
}

func (osds *Osds) fakeCreate() (bool, error) {
	diskIDs := osds.getIntegerListValue(osds.DiskIDs)
	osdIDs := []int64{}
	for range diskIDs {
		osdIDs = append(osdIDs, rand.Int63())
	}
	osds.repr = osdIDs
	return true, nil
}

func (osds *Osds) getResource(diskIDs []int64) (diskMap map[int64]int64, err error) {
	body, err := osds.CallResourceAPI(utils.ListAPIName, nil, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to list osds")
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
		return nil, errors.Annotate(err, "parse list osds data")
	}

	diskMap = make(map[int64]int64)
	for _, osdResp := range records.Osds {
		for _, diskID := range diskIDs {
			if osdResp.Disk.ID == diskID {
				diskMap[diskID] = osdResp.ID
			}
		}
	}
	return diskMap, nil
}

// Create create the resource
func (osds *Osds) Create() (created bool, err error) {
	if osds.DiskIDs == nil {
		return false, errors.Errorf("HostIDs is required")
	}
	if config.DryRun {
		return osds.fakeCreate()
	}

	diskIDs := osds.getIntegerListValue(osds.DiskIDs)
	if len(diskIDs) == 0 {
		osds.repr = []int64{}
		log.Printf("Skip creating osds with zero disks")
		return true, nil
	}
	diskMap, err := osds.getResource(diskIDs)
	if err != nil {
		return false, errors.Annotatef(err, "get osds with disks %+v", diskIDs)
	}
	partitionIDs := make([]int64, 0, len(diskIDs))
	if osds.PartitionIDs != nil {
		partitionIDs = osds.getIntegerListValue(osds.PartitionIDs)
	}
	role := utils.PoolOsdRoleData
	if osds.Role != nil {
		role = osds.getStringValue(osds.Role)
	}
	log.Printf("try to create %d %s osds using %d cache disks", len(diskIDs), role, len(partitionIDs))

	osdIDs := []int64{}
	for index, diskID := range diskIDs {
		osdID, ok := diskMap[diskID]
		if ok {
			osdIDs = append(osdIDs, osdID)
			continue
		}

		req := new(OsdCreateReq)
		req.Osd.DiskID = diskID
		if index < len(partitionIDs) {
			req.Osd.PartitionID = partitionIDs[index]
		}
		req.Osd.Role = role
		if osds.OmapByte != nil {
			req.Osd.OmapByte = osds.getIntegerValue(osds.OmapByte)
		}

		body, err := osds.CallCreateAPI(req, nil)
		if err != nil {
			return false, errors.Annotatef(err, "create osd with disk %d", diskID)
		}
		id, _, err := osds.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}

		osdIDs = append(osdIDs, id.(int64))
	}

	osds.repr = osdIDs
	osds.creatingOsdIDs = osdIDs
	return false, nil
}

// IsCreated if the resource has been created
func (osds *Osds) IsCreated() (created bool, err error) {
	creatingOsdIDs := []int64{}
	getReqIdentify, err := settings.GetSetting(osds.GetType(), utils.GetReqIdentify)
	if err != nil {
		return false, errors.Trace(err)
	}
	for _, osdID := range osds.creatingOsdIDs {
		pathParam := map[string]string{getReqIdentify: fmt.Sprintf("%d", osdID)}
		body, err := osds.CallGetAPI(pathParam)
		if err != nil {
			return false, errors.Annotatef(err, "get osd with id %d", osdID)
		}
		_, status, err := osds.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}
		created, err = osds.checkStatus(status)
		if err != nil {
			return false, errors.Trace(err)
		}
		if created {
			log.Printf("item osd %d is created", osdID)
		} else {
			creatingOsdIDs = append(creatingOsdIDs, osdID)
		}
	}

	if len(creatingOsdIDs) == 0 {
		return true, nil
	}

	osds.creatingOsdIDs = creatingOsdIDs
	return false, nil
}
