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

// Partitions resource
type Partitions struct {
	ResourceBase
	// TODO(wuhao): support partition here
	DiskIDs    *parser.IntegerListExpr
	NumPerDisk *parser.IntegerExpr

	cachingDiskIDs []int64
	partitionIDs   []int64
}

// Init inits resource instance
func (partitions *Partitions) Init(stack utils.StackInterface) {
	partitions.ResourceBase.Init(stack)
	partitions.setDelegate(partitions)
}

// GetType return resource type
func (partitions *Partitions) GetType() string {
	return utils.ResourcePartitions
}

// CheckInterval return check interval
func (partitions *Partitions) CheckInterval() int {
	return 5 * len(partitions.cachingDiskIDs)
}

// IsReady check if the formation args are ready
func (partitions *Partitions) IsReady() (ready bool) {
	if !partitions.isReady(partitions.DiskIDs) ||
		!partitions.isReady(partitions.NumPerDisk) {
		return false
	}

	return true
}

func (partitions *Partitions) fakeCreate() (bool, error) {
	for i := 0; i < int(partitions.NumPerDisk.Literal); i++ {
		partitions.partitionIDs = append(partitions.partitionIDs, rand.Int63())
	}
	partitions.repr = partitions.partitionIDs
	return true, nil
}

// Create create the resource
func (partitions *Partitions) Create() (created bool, err error) {
	if partitions.DiskIDs == nil {
		return false, errors.Errorf("HostIDs is required")
	}

	if config.DryRun {
		return partitions.fakeCreate()
	}

	var numPerDisk int64 = 1
	if partitions.NumPerDisk != nil {
		numPerDisk = partitions.getIntegerValue(partitions.NumPerDisk)
	}

	getReqIdentifyKey, err := settings.GetSetting(partitions.GetType(), utils.GetReqIdentify)
	if err != nil {
		return false, errors.Trace(err)
	}
	diskIDs := partitions.getIntegerListValue(partitions.DiskIDs)
	for _, diskID := range diskIDs {
		pathParam := map[string]string{getReqIdentifyKey: fmt.Sprintf("%d", diskID)}
		queryParam := map[string]string{"num": fmt.Sprintf("%d", numPerDisk)}
		_, err = partitions.CallCreateAPI(nil, pathParam, queryParam)
		if err != nil {
			return false, errors.Annotatef(err, "create partition with disk %d", diskID)
		}
	}

	partitions.cachingDiskIDs = diskIDs
	partitions.partitionIDs = make([]int64, 0, int(numPerDisk)*len(diskIDs))
	return false, nil
}

// IsCreated if the resource has been created
func (partitions *Partitions) IsCreated() (created bool, err error) {
	cachingDiskIDs := []int64{}
	getReqIdentifyKey, err := settings.GetSetting(partitions.GetType(), utils.GetReqIdentify)
	if err != nil {
		return false, errors.Trace(err)
	}
	for _, diskID := range partitions.cachingDiskIDs {
		pathParam := map[string]string{getReqIdentifyKey: fmt.Sprintf("%d", diskID)}
		body, err := partitions.CallGetAPI(pathParam)
		if err != nil {
			return false, errors.Trace(err)
		}
		resp := &struct {
			Disk struct {
				ID           int64  `json:"id"`
				Status       string `json:"status"`
				ActionStatus string `json:"action_status"`
				Partitions   []struct {
					ID int64 `json:"id"`
				} `json:"partitions"`
			} `json:"disk"`
		}{}
		if err = json.Unmarshal(body, resp); err != nil {
			return false, errors.Trace(err)
		}
		if resp.Disk.ActionStatus != "" {
			created, err = partitions.checkStatus(resp.Disk.ActionStatus)
			if err != nil {
				return false, errors.Trace(err)
			}
		}
		if created {
			for _, partitionResp := range resp.Disk.Partitions {
				partitions.partitionIDs = append(partitions.partitionIDs, partitionResp.ID)
			}
			log.Printf("partitions are created on disk %d", diskID)
		} else {
			cachingDiskIDs = append(cachingDiskIDs, diskID)
		}
	}

	if len(cachingDiskIDs) == 0 {
		partitions.repr = partitions.partitionIDs
		return true, nil
	}

	partitions.cachingDiskIDs = cachingDiskIDs
	return false, nil
}
