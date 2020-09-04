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

// BlockVolumes resource
type BlockVolumes struct {
	ResourceBase

	BlockSnapshotID     *parser.IntegerExpr
	Description         *parser.StringExpr
	Flattened           *parser.BoolExpr
	Format              *parser.IntegerExpr
	PerformancePriority *parser.IntegerExpr
	PoolID              *parser.IntegerExpr
	Qos                 *BlockVolumeQos
	QosEnabled          *parser.BoolExpr
	Size                *parser.IntegerExpr

	Names                  *parser.StringListExpr
	Prefix                 *parser.StringExpr
	Num                    *parser.IntegerExpr
	creatingBlockVolumeIDs []int64
}

// Init inits resource instance
func (volumes *BlockVolumes) Init(stack utils.StackInterface) {
	volumes.ResourceBase.Init(stack)
	volumes.setDelegate(volumes)
}

// GetType return resource type
func (volumes *BlockVolumes) GetType() string {
	return utils.ResourceBlockVolumes
}

// CheckInterval get check interval
func (volumes *BlockVolumes) CheckInterval() int {
	return 1 * len(volumes.creatingBlockVolumeIDs)
}

// IsReady check if the formation args are ready
func (volumes *BlockVolumes) IsReady() (ready bool) {
	if !volumes.isReady(volumes.BlockSnapshotID) ||
		!volumes.isReady(volumes.Description) ||
		!volumes.isReady(volumes.Flattened) ||
		!volumes.isReady(volumes.Format) ||
		!volumes.isReady(volumes.PerformancePriority) ||
		!volumes.isReady(volumes.PoolID) ||
		!volumes.isReady(volumes.Flattened) ||
		!volumes.isReady(volumes.QosEnabled) ||
		!volumes.isReady(volumes.Size) ||
		!volumes.isReady(volumes.Names) ||
		!volumes.isReady(volumes.Prefix) ||
		!volumes.isReady(volumes.Num) {
		return false
	}

	qos := volumes.Qos
	if qos != nil {
		if !volumes.isReady(qos.BurstTotalBw) ||
			!volumes.isReady(qos.BurstTotalIops) ||
			!volumes.isReady(qos.MaxTotalBw) ||
			!volumes.isReady(qos.MaxTotalIops) {
			return false
		}
	}

	return true
}

func (volumes *BlockVolumes) getNames() (names []string, err error) {
	names = []string{}
	if volumes.Names != nil {
		names = volumes.getStringListValue(volumes.Names)
	}
	if volumes.Prefix != nil && volumes.Num != nil {
		num := volumes.getIntegerValue(volumes.Num)
		prefix := volumes.getStringValue(volumes.Prefix)
		for i := int64(1); i <= num; i++ {
			name := fmt.Sprintf("%s-%d", prefix, i)
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		return nil, errors.Errorf("Names or Prefix with Num are required")
	}
	return names, nil
}

func (volumes *BlockVolumes) fakeCreate(names []string) (bool, error) {
	volumeIDs := []int64{}
	for range names {
		volumeIDs = append(volumeIDs, rand.Int63())
	}
	volumes.repr = volumeIDs
	return true, nil
}

func (volumes *BlockVolumes) getResource(names []string) (volumeMap map[string]int64, err error) {
	body, err := volumes.CallResourceAPI(utils.ListAPIName, nil, nil)
	if err != nil {
		return nil, errors.Annotatef(err, "list block volumes")
	}
	records := &struct {
		Volumes []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"block_volumes"`
	}{}
	if err = json.Unmarshal(body, records); err != nil {
		return nil, errors.Trace(err)
	}

	volumeMap = make(map[string]int64)
	for _, volumeResp := range records.Volumes {
		for _, name := range names {
			if volumeResp.Name == name {
				volumeMap[name] = volumeResp.ID
			}
		}
	}
	return volumeMap, nil
}

// Create create the resource
func (volumes *BlockVolumes) Create() (created bool, err error) {
	names, err := volumes.getNames()
	if err != nil {
		err = errors.Annotate(err, "failed to generate names for block volumes")
		return
	}
	if config.DryRun {
		return volumes.fakeCreate(names)
	}

	volumeMap, err := volumes.getResource(names)
	if err != nil {
		err = errors.Annotatef(err, "failed to get block volumes %+v", names)
		return
	}
	log.Printf("try to create block volumes %+v", names)

	volumeIDs := []int64{}
	req := new(VolumeCreateReq)
	volumeInfo := &req.Volume
	if volumes.BlockSnapshotID != nil {
		volumeInfo.BlockSnapshotID = volumes.getIntegerValue(volumes.BlockSnapshotID)
	}
	if volumes.Description != nil {
		volumeInfo.Description = volumes.getStringValue(volumes.Description)
	}
	if volumes.Flattened != nil {
		volumeInfo.Flattened = volumes.getBoolValue(volumes.Flattened)
	}
	if volumes.Format != nil {
		volumeInfo.Format = volumes.getIntegerValue(volumes.Format)
	}
	if volumes.PerformancePriority != nil {
		volumeInfo.PerformancePriority = volumes.getIntegerValue(volumes.PerformancePriority)
	}
	if volumes.PoolID != nil {
		volumeInfo.PoolID = volumes.getIntegerValue(volumes.PoolID)
	}
	if volumes.QosEnabled != nil {
		volumeInfo.QosEnabled = volumes.getBoolValue(volumes.QosEnabled)
	}
	if volumes.Size != nil {
		volumeInfo.Size = volumes.getIntegerValue(volumes.Size)
	}
	qos := volumes.Qos
	if qos != nil {
		qosReq := new(volumeQosSpec)
		if qos.BurstTotalBw != nil {
			qosReq.BurstTotalBw = volumes.getIntegerValue(qos.BurstTotalBw)
		}
		if qos.BurstTotalIops != nil {
			qosReq.BurstTotalIops = volumes.getIntegerValue(qos.BurstTotalIops)
		}
		if qos.MaxTotalBw != nil {
			qosReq.MaxTotalBw = volumes.getIntegerValue(qos.MaxTotalBw)
		}
		if qos.MaxTotalIops != nil {
			qosReq.MaxTotalIops = volumes.getIntegerValue(qos.MaxTotalIops)
		}
		volumeInfo.Qos = qosReq
	}

	for _, name := range names {
		volumeID, ok := volumeMap[name]
		if ok {
			volumeIDs = append(volumeIDs, volumeID)
			continue
		}

		volumeInfo.Name = name
		body, err := volumes.CallCreateAPI(req, nil)
		if err != nil {
			return false, errors.Annotatef(err, "create volume %s", name)
		}
		id, _, err := volumes.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}
		volumeIDs = append(volumeIDs, id.(int64))
	}

	volumes.repr = volumeIDs
	volumes.creatingBlockVolumeIDs = volumeIDs
	return false, nil
}

// IsCreated if the resource has been created
func (volumes *BlockVolumes) IsCreated() (created bool, err error) {
	creatingBlockVolumeIDs := []int64{}
	for _, volumeID := range volumes.creatingBlockVolumeIDs {
		identifyKey, err := settings.GetSetting(volumes.GetType(), utils.GetReqIdentify)
		if err != nil {
			return false, errors.Trace(err)
		}
		pathParam := map[string]string{identifyKey: fmt.Sprintf("%d", volumeID)}
		body, err := volumes.CallGetAPI(pathParam)
		if err != nil {
			return false, errors.Annotatef(err, "get block volume")
		}
		_, status, err := volumes.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}
		created, err = volumes.checkStatus(status)
		if err != nil {
			return false, errors.Trace(err)
		}
		if created {
			log.Printf("item block volume %d is created", volumeID)
		} else {
			creatingBlockVolumeIDs = append(creatingBlockVolumeIDs, volumeID)
		}
	}

	if len(creatingBlockVolumeIDs) == 0 {
		return true, nil
	}
	volumes.creatingBlockVolumeIDs = creatingBlockVolumeIDs
	return false, nil
}
