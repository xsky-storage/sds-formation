package formation

import (
	"fmt"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// DiskUpdateReq defines disk update request
type DiskUpdateReq struct {
	Disk struct {
		DiskType       string `json:"disk_type,omitempty"`
		LightingStatus string `json:"lighting_status,omitempty"`
		PowerSafe      bool   `json:"power_safe,omitempty"`
	} `json:"disk"`
}

// DiskListUpdate resource
type DiskListUpdate struct {
	ResourceBase
	DiskType       *parser.StringExpr
	LightingStatus *parser.StringExpr
}

// Init inits resource instance
func (diskList *DiskListUpdate) Init(stack utils.StackInterface) {
	diskList.ResourceBase.Init(stack)
	diskList.setDelegate(diskList)
}

// GetType return resource type
func (diskList *DiskListUpdate) GetType() string {
	return utils.ResourceDiskList
}

// IsReady check if the formation args are ready
func (diskList *DiskListUpdate) IsReady() (ready bool) {
	if !diskList.isReady(diskList.DiskType) {
		return false
	}
	return true
}

// Update update the resource
func (diskList *DiskListUpdate) Update(repr interface{}) (updated bool, err error) {
	diskIDs, ok := repr.([]int64)
	if !ok {
		return false, errors.Errorf("unexpected repr!!!")
	}
	if config.DryRun {
		diskList.repr = diskIDs
		return true, nil
	}

	req := new(DiskUpdateReq)
	if diskList.DiskType != nil {
		req.Disk.DiskType = diskList.getStringValue(diskList.DiskType)
	}
	if diskList.LightingStatus != nil {
		req.Disk.LightingStatus = diskList.getStringValue(diskList.LightingStatus)
	}

	reqIdentifyKey, err := settings.GetSetting(diskList.GetType(), utils.GetReqIdentify)
	if err != nil {
		return false, errors.Trace(err)
	}
	for _, diskID := range diskIDs {
		pathParam := map[string]string{reqIdentifyKey: fmt.Sprintf("%d", diskID)}
		_, err := diskList.CallResourceAPI(utils.UpdateAPIName, req, pathParam)
		if err != nil {
			return false, errors.Annotatef(err, "update disk %d", diskID)
		}
	}

	diskList.repr = diskIDs
	return true, nil
}
