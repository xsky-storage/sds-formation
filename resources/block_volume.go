package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type volumeQosSpec struct {
	BurstTotalBw int64 `json:"burst_total_bw,omitempty"`

	BurstTotalIops int64 `json:"burst_total_iops,omitempty"`

	MaxTotalBw int64 `json:"max_total_bw,omitempty"`

	MaxTotalIops int64 `json:"max_total_iops,omitempty"`
}

// VolumeCreateReq defines block volume create request
type VolumeCreateReq struct {
	Volume struct {

		// id of related block volume snapshot
		BlockSnapshotID int64 `json:"block_snapshot_id,omitempty"`

		// description of volume
		Description string `json:"description,omitempty"`

		// flatten or not flatten
		Flattened bool `json:"flattened,omitempty"`

		// volume format: { 128 | 129 (advanced) }, default 128
		Format int64 `json:"format,omitempty"`

		// name of volume
		Name string `json:"name"`

		// performance priority: { 0 | 1 }, default 0
		PerformancePriority int64 `json:"performance_priority,omitempty"`

		// id of pool belonged to
		PoolID int64 `json:"pool_id"`

		// qos of volume
		Qos *volumeQosSpec `json:"qos,omitempty"`

		// enable or disable the qos
		QosEnabled bool `json:"qos_enabled,omitempty"`

		// replication remote cluster fsid
		RemoteClusterFsID string `json:"remote_cluster_fs_id,omitempty"`

		// replication peer pool
		ReplicationPool string `json:"replication_pool,omitempty"`

		// replication peer pool id
		ReplicationPoolID int64 `json:"replication_pool_id,omitempty"`

		// replication peer pool name
		ReplicationPoolName string `json:"replication_pool_name,omitempty"`

		// replication version
		ReplicationVersion int64 `json:"replication_version,omitempty"`

		// replication peer volume
		ReplicationVolume string `json:"replication_volume,omitempty"`

		// replication peer volume id
		ReplicationVolumeID int64 `json:"replication_volume_id,omitempty"`

		// replication peer volume name
		ReplicationVolumeName string `json:"replication_volume_name,omitempty"`

		// size of volume
		Size int64 `json:"size,omitempty"`

		// volume sn, used when creating replication volume
		Sn string `json:"sn,omitempty"`

		// snapshot replication peer pool
		SnapshotReplicationPool string `json:"snapshot_replication_pool,omitempty"`

		// snapshot replication peer pool id
		SnapshotReplicationPoolID int64 `json:"snapshot_replication_pool_id,omitempty"`

		// snapshot replication peer volume
		SnapshotReplicationVolume string `json:"snapshot_replication_volume,omitempty"`

		// snapshot replication peer volume id
		SnapshotReplicationVolumeID int64 `json:"snapshot_replication_volume_id,omitempty"`
	} `json:"block_volume"`
}

// BlockVolumeQos resource
type BlockVolumeQos struct {
	BurstTotalBw   *parser.IntegerExpr
	BurstTotalIops *parser.IntegerExpr
	MaxTotalBw     *parser.IntegerExpr
	MaxTotalIops   *parser.IntegerExpr
}

// BlockVolume resource
type BlockVolume struct {
	ResourceBase
	BlockSnapshotID     *parser.IntegerExpr
	Description         *parser.StringExpr
	Flattened           *parser.BoolExpr
	Format              *parser.IntegerExpr
	Name                *parser.StringExpr
	PerformancePriority *parser.IntegerExpr
	PoolID              *parser.IntegerExpr
	Qos                 *BlockVolumeQos
	QosEnabled          *parser.BoolExpr
	Size                *parser.IntegerExpr
}

// Init inits resource instance
func (volume *BlockVolume) Init(stack utils.StackInterface) {
	volume.ResourceBase.Init(stack)
	volume.setDelegate(volume)
}

// GetType return resource type
func (volume *BlockVolume) GetType() string {
	return utils.ResourceBlockVolume
}

// IsReady check if the formation args are ready
func (volume *BlockVolume) IsReady() (ready bool) {
	if !volume.isReady(volume.BlockSnapshotID) ||
		!volume.isReady(volume.Description) ||
		!volume.isReady(volume.Flattened) ||
		!volume.isReady(volume.Format) ||
		!volume.isReady(volume.Name) ||
		!volume.isReady(volume.PerformancePriority) ||
		!volume.isReady(volume.PoolID) ||
		!volume.isReady(volume.Flattened) ||
		!volume.isReady(volume.QosEnabled) ||
		!volume.isReady(volume.Size) {
		return false
	}

	qos := volume.Qos
	if qos != nil {
		if !volume.isReady(qos.BurstTotalBw) ||
			!volume.isReady(qos.BurstTotalIops) ||
			!volume.isReady(qos.MaxTotalBw) ||
			!volume.isReady(qos.MaxTotalIops) {
			return false
		}
	}

	return true
}

func (volume *BlockVolume) fakeCreate() (bool, error) {
	volume.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (volume *BlockVolume) Create() (created bool, err error) {
	if volume.Name == nil {
		return false, errors.Errorf("Name is required for resource %s", volume.GetType())
	}
	if config.DryRun {
		return volume.fakeCreate()
	}

	name := volume.getStringValue(volume.Name)
	resourceID, err := volume.getResourceByName(name)
	if err != nil {
		return false, errors.Annotatef(err, "get volume %s", name)
	}
	if resourceID != nil {
		volume.repr = resourceID
		return false, nil
	}

	req := new(VolumeCreateReq)
	volumeInfo := &req.Volume
	if volume.BlockSnapshotID != nil {
		volumeInfo.BlockSnapshotID = volume.getIntegerValue(volume.BlockSnapshotID)
	}
	if volume.Description != nil {
		volumeInfo.Description = volume.getStringValue(volume.Description)
	}
	if volume.Flattened != nil {
		volumeInfo.Flattened = volume.getBoolValue(volume.Flattened)
	}
	if volume.Format != nil {
		volumeInfo.Format = volume.getIntegerValue(volume.Format)
	}
	if volume.Name != nil {
		volumeInfo.Name = volume.getStringValue(volume.Name)
	}
	if volume.PerformancePriority != nil {
		volumeInfo.PerformancePriority = volume.getIntegerValue(volume.PerformancePriority)
	}
	if volume.PoolID != nil {
		volumeInfo.PoolID = volume.getIntegerValue(volume.PoolID)
	}
	if volume.QosEnabled != nil {
		volumeInfo.QosEnabled = volume.getBoolValue(volume.QosEnabled)
	}
	if volume.Size != nil {
		volumeInfo.Size = volume.getIntegerValue(volume.Size)
	}
	qos := volume.Qos
	if qos != nil {
		qosReq := new(volumeQosSpec)
		if qos.BurstTotalBw != nil {
			qosReq.BurstTotalBw = volume.getIntegerValue(qos.BurstTotalBw)
		}
		if qos.BurstTotalIops != nil {
			qosReq.BurstTotalIops = volume.getIntegerValue(qos.BurstTotalIops)
		}
		if qos.MaxTotalBw != nil {
			qosReq.MaxTotalBw = volume.getIntegerValue(qos.MaxTotalBw)
		}
		if qos.MaxTotalIops != nil {
			qosReq.MaxTotalIops = volume.getIntegerValue(qos.MaxTotalIops)
		}
		volumeInfo.Qos = qosReq
	}

	body, err := volume.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create volume %s", name)
	}

	id, status, err := volume.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	volume.repr = id
	created, err = volume.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return created, nil
}
