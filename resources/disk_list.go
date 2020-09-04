package formation

import (
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type partition struct {
	Create time.Time `json:"create,omitempty"`

	Disk *DiskRecord `json:"disk,omitempty"`

	ID int64 `json:"id,omitempty"`

	Path string `json:"path,omitempty"`

	Size int64 `json:"size,omitempty"`

	Update time.Time `json:"update,omitempty"`

	Used bool `json:"used,omitempty"`

	UUID string `json:"uuid,omitempty"`
}

type hostNestview struct {
	AdminIP string `json:"admin_ip"`

	ID int64 `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

type smartAttr struct {
	AttrID int64 `json:"attr_id,omitempty"`

	Create time.Time `json:"create,omitempty"`

	Flag string `json:"flag,omitempty"`

	ID int64 `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	RawValue string `json:"raw_value,omitempty"`

	Status string `json:"status,omitempty"`

	Thresh string `json:"thresh,omitempty"`

	Type string `json:"type,omitempty"`

	Value string `json:"value,omitempty"`

	WhenFailed string `json:"when_failed,omitempty"`

	Worst string `json:"worst,omitempty"`
}

// DiskRecord defines disk info from api response
type DiskRecord struct {
	ActionStatus string `json:"action_status,omitempty"`

	// size of disk
	Bytes int64 `json:"bytes,omitempty"`

	CacheCreate time.Time `json:"cache_create,omitempty"`

	Create time.Time `json:"create,omitempty"`

	Device string `json:"device,omitempty"`

	DiskType string `json:"disk_type,omitempty"`

	DriverType string `json:"driver_type,omitempty"`

	EnclosureID string `json:"enclosure_id,omitempty"`

	Host *hostNestview `json:"host,omitempty"`

	ID int64 `json:"id,omitempty"`

	// used as cache disk
	IsCache bool `json:"is_cache,omitempty"`

	// used as root disk
	IsRoot bool `json:"is_root,omitempty"`

	LightingStatus string `json:"lighting_status,omitempty"`

	Model string `json:"model,omitempty"`

	PartitionNum int64 `json:"partition_num,omitempty"`

	Partitions []partition `json:"partitions,omitempty"`

	PowerSafe bool `json:"power_safe,omitempty"`

	RotationRate string `json:"rotation_rate,omitempty"`

	Rotational bool `json:"rotational,omitempty"`

	Serial string `json:"serial,omitempty"`

	SlotID string `json:"slot_id,omitempty"`

	SmartAttrs []smartAttr `json:"smart_attrs,omitempty"`

	SsdLifeLeft int64 `json:"ssd_life_left,omitempty"`

	Status string `json:"status,omitempty"`

	Update time.Time `json:"update,omitempty"`

	Used bool `json:"used,omitempty"`

	Wwid string `json:"wwid,omitempty"`
}

// DisksResp defines disk list api response
type DisksResp struct {
	Disks []*DiskRecord `json:"disks"`
}

// DiskList resource
type DiskList struct {
	ResourceBase
	Used       *parser.BoolExpr
	Device     *parser.StringExpr
	DiskType   *parser.StringExpr
	Model      *parser.StringExpr
	MinSizeGB  *parser.IntegerExpr
	MaxSizeGB  *parser.IntegerExpr
	HostIDs    *parser.IntegerListExpr
	Num        *parser.IntegerExpr
	NumPerHost *parser.IntegerExpr
	IsCache    *parser.BoolExpr
	Status     *parser.StringExpr
	WWID       *parser.StringExpr
}

// Init inits resource instance
func (diskList *DiskList) Init(stack utils.StackInterface) {
	diskList.ResourceBase.Init(stack)
	diskList.setDelegate(diskList)
}

// GetType return resource type
func (diskList *DiskList) GetType() string {
	return utils.ResourceDiskList
}

// IsReady check if the formation args are ready
func (diskList *DiskList) IsReady() (ready bool) {
	if !diskList.isReady(diskList.Used) ||
		!diskList.isReady(diskList.Device) ||
		!diskList.isReady(diskList.DiskType) ||
		!diskList.isReady(diskList.Model) ||
		!diskList.isReady(diskList.MinSizeGB) ||
		!diskList.isReady(diskList.MaxSizeGB) ||
		!diskList.isReady(diskList.HostIDs) ||
		!diskList.isReady(diskList.IsCache) ||
		!diskList.isReady(diskList.Status) ||
		!diskList.isReady(diskList.WWID) {

		return false
	}

	return true
}

func (diskList *DiskList) fakeCreate() (bool, error) {
	diskIDs := []int64{}
	if diskList.HostIDs == nil {
		diskNum := 2
		if diskList.Num != nil {
			diskNum = int(diskList.getIntegerValue(diskList.Num))
		}
		for i := 0; i < diskNum; i++ {
			diskIDs = append(diskIDs, rand.Int63())
		}
		diskList.repr = diskIDs
		return true, nil
	}

	hostIDs := diskList.getIntegerListValue(diskList.HostIDs)
	for range hostIDs {
		diskIDs = append(diskIDs, rand.Int63(), rand.Int63())
	}
	diskList.repr = diskIDs
	return true, nil
}

// Create create the resource
func (diskList *DiskList) Create() (created bool, err error) {
	if config.DryRun {
		return diskList.fakeCreate()
	}

	diskIDs := []int64{}
	if diskList.HostIDs == nil {
		diskIDs, err = diskList.getDiskIDs()
		if err != nil {
			err = errors.Trace(err)
			return
		}
	} else {
		hostIDs := diskList.getIntegerListValue(diskList.HostIDs)
		for _, hostID := range hostIDs {
			ids, e := diskList.getDiskIDs(hostID)
			if e != nil {
				return false, e
			}
			diskIDs = append(diskIDs, ids...)
		}
	}

	if diskList.Num != nil {
		num := int(diskList.getIntegerValue(diskList.Num))
		if len(diskIDs) >= num {
			diskIDs = diskIDs[:num]
		} else {
			return false, errors.Errorf("failed to get %d valid disks", num)
		}
	}
	diskList.repr = diskIDs
	return true, nil
}

func (diskList *DiskList) getDiskIDs(args ...int64) (diskIDs []int64, err error) {
	argMap := make(map[string]string)
	if diskList.Used != nil {
		used, err := diskList.getValString(diskList.getBoolValue(diskList.Used))
		if err != nil {
			return nil, errors.Annotatef(err, "parse used flag")
		}
		argMap["used"] = used
	}
	if len(args) > 0 {
		hostID, err := diskList.getValString(args[0])
		if err != nil {
			return nil, errors.Annotatef(err, "parse host id")
		}
		argMap["host_id"] = hostID
	}
	// unlimit
	argMap["limit"] = "-1"

	body, err := diskList.CallResourceAPI(utils.ListAPIName, nil, nil, argMap)
	if err != nil {
		return nil, errors.Annotatef(err, "list disks")
	}
	disksResp := new(DisksResp)
	if json.Unmarshal(body, disksResp); err != nil {
		return nil, errors.Annotatef(err, "parse response body")
	}

	disksPerHostMap := map[int64]int64{}
	var diskPerHost int64 = 0
	if diskList.NumPerHost != nil {
		diskPerHost = diskList.getIntegerValue(diskList.NumPerHost)
	}
	diskIDs = []int64{}
	for _, disk := range disksResp.Disks {

		if diskList.Device != nil && disk.Device != diskList.getStringValue(diskList.Device) {
			continue
		}
		if diskList.DiskType != nil && disk.DiskType != diskList.getStringValue(diskList.DiskType) {
			continue
		}
		if diskList.Model != nil &&
			!strings.Contains(disk.Model, diskList.getStringValue(diskList.Model)) {

			continue
		}
		if diskList.MinSizeGB != nil &&
			disk.Bytes < diskList.getIntegerValue(diskList.MinSizeGB)*1024*1024*1024 {

			continue
		}
		if diskList.MaxSizeGB != nil &&
			disk.Bytes > diskList.getIntegerValue(diskList.MaxSizeGB)*1024*1024*1024 {

			continue
		}
		if diskList.IsCache != nil && disk.IsCache != diskList.getBoolValue(diskList.IsCache) {
			continue
		}
		if diskList.Status != nil && disk.Status != diskList.getStringValue(diskList.Status) {
			continue
		}
		if diskList.WWID != nil &&
			!strings.Contains(disk.Wwid, diskList.getStringValue(diskList.WWID)) {

			continue
		}

		if disk.Status != utils.StatusActive {
			log.Printf("disk %d in status %s is skipped", disk.ID, disk.Status)
		} else {
			if diskList.NumPerHost != nil {
				if disk.Host == nil || disksPerHostMap[disk.Host.ID] >= diskPerHost {
					continue
				}
				disksPerHostMap[disk.Host.ID]++
			}
			diskIDs = append(diskIDs, disk.ID)
		}
	}
	return diskIDs, nil
}
