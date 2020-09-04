package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSFolderCreateReq defines request for file storage folder creation
type FSFolderCreateReq struct {
	Folder struct {
		// name of folder
		Name string `json:"name" required:"true"`
		// description of folder
		Description string `json:"description,omitempty"`
		// id of pool
		PoolID int64 `json:"pool_id"`
		// size of folder
		Size int64 `json:"size"`
		// qos of volume
		Qos *volumeQosSpec `json:"qos,omitempty"`
		// enable or disable the qos
		QosEnabled bool `json:"qos_enabled,omitempty"`
		// file storage snapshot id
		FSSnapshotID int64 `json:"fs_snapshot_id,omitempty"`
		// flatten or not flatten
		Flattened bool `json:"flattened,omitempty"`
		// file storage folder
	} `json:"fs_folder" required:"true"`
}

// FSFolder resource
type FSFolder struct {
	ResourceBase

	Name         *parser.StringExpr
	Description  *parser.StringExpr
	PoolID       *parser.IntegerExpr
	Size         *parser.IntegerExpr
	Qos          *BlockVolumeQos
	QosEnabled   *parser.BoolExpr
	FSSnapshotID *parser.IntegerExpr
	Flattened    *parser.BoolExpr
}

// Init inits resource instance
func (folder *FSFolder) Init(stack utils.StackInterface) {
	folder.ResourceBase.Init(stack)
	folder.setDelegate(folder)
}

// GetType return resource type
func (folder *FSFolder) GetType() string {
	return utils.ResourceFSFolder
}

// IsReady check if the formation args are ready
func (folder *FSFolder) IsReady() (ready bool) {
	if !folder.isReady(folder.Name) ||
		!folder.isReady(folder.Description) ||
		!folder.isReady(folder.PoolID) ||
		!folder.isReady(folder.Size) ||
		!folder.isReady(folder.QosEnabled) ||
		!folder.isReady(folder.FSSnapshotID) ||
		!folder.isReady(folder.Flattened) {
		return false
	}

	qos := folder.Qos
	if qos != nil {
		if !folder.isReady(qos.BurstTotalBw) ||
			!folder.isReady(qos.BurstTotalIops) ||
			!folder.isReady(qos.MaxTotalBw) ||
			!folder.isReady(qos.MaxTotalIops) {
			return false
		}
	}

	return true
}

func (folder *FSFolder) fakeCreate() (bool, error) {
	folder.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (folder *FSFolder) Create() (created bool, err error) {
	if folder.Name == nil {
		err = errors.Errorf("Name is required for resource %s", folder.GetType())
		return
	}
	if config.DryRun {
		return folder.fakeCreate()
	}

	name := folder.getStringValue(folder.Name)
	resourceID, err := folder.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		folder.repr = resourceID
		return false, nil
	}

	req := new(FSFolderCreateReq)
	folderInfo := &req.Folder
	if folder.Name != nil {
		folderInfo.Name = folder.getStringValue(folder.Name)
	}
	if folder.Description != nil {
		folderInfo.Description = folder.getStringValue(folder.Description)
	}
	if folder.PoolID != nil {
		folderInfo.PoolID = folder.getIntegerValue(folder.PoolID)
	}
	if folder.Size != nil {
		folderInfo.Size = folder.getIntegerValue(folder.Size)
	}
	if folder.QosEnabled != nil {
		folderInfo.QosEnabled = folder.getBoolValue(folder.QosEnabled)
	}
	if folder.FSSnapshotID != nil {
		folderInfo.FSSnapshotID = folder.getIntegerValue(folder.FSSnapshotID)
	}
	if folder.Flattened != nil {
		folderInfo.Flattened = folder.getBoolValue(folder.Flattened)
	}

	qos := folder.Qos
	if qos != nil {
		qosReq := new(volumeQosSpec)
		if qos.BurstTotalBw != nil {
			qosReq.BurstTotalBw = folder.getIntegerValue(qos.BurstTotalBw)
		}
		if qos.BurstTotalIops != nil {
			qosReq.BurstTotalIops = folder.getIntegerValue(qos.BurstTotalIops)
		}
		if qos.MaxTotalBw != nil {
			qosReq.MaxTotalBw = folder.getIntegerValue(qos.MaxTotalBw)
		}
		if qos.MaxTotalIops != nil {
			qosReq.MaxTotalIops = folder.getIntegerValue(qos.MaxTotalIops)
		}
		folderInfo.Qos = qosReq
	}

	body, err := folder.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs folder %s", name)
	}

	id, status, err := folder.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	folder.repr = id
	created, err = folder.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
