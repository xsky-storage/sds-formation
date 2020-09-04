package formation

import (
	"encoding/json"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// MappingGroupCreateReq defines mapping group create request
type MappingGroupCreateReq struct {
	MappingGroup struct {
		AccessPathID int64 `json:"access_path_id,omitempty"`

		BlockVolumeIds []int64 `json:"block_volume_ids,omitempty"`

		ClientGroupID int64 `json:"client_group_id,omitempty"`
	} `json:"mapping_group"`
}

// MappingGroup resource
type MappingGroup struct {
	ResourceBase

	AccessPathID   *parser.IntegerExpr
	BlockVolumeIDs *parser.IntegerListExpr
	ClientGroupID  *parser.IntegerExpr
}

// Init inits resource instance
func (mappingGroup *MappingGroup) Init(stack utils.StackInterface) {
	mappingGroup.ResourceBase.Init(stack)
	mappingGroup.setDelegate(mappingGroup)
}

// GetType return resource type
func (mappingGroup *MappingGroup) GetType() string {
	return utils.ResourceMappingGroup
}

// IsReady check if the formation args are ready
func (mappingGroup *MappingGroup) IsReady() (ready bool) {
	if !mappingGroup.isReady(mappingGroup.AccessPathID) ||
		!mappingGroup.isReady(mappingGroup.BlockVolumeIDs) ||
		!mappingGroup.isReady(mappingGroup.ClientGroupID) {
		return false
	}

	return true
}

func (mappingGroup *MappingGroup) fakeCreate() (bool, error) {
	mappingGroup.repr = rand.Int63()
	return true, nil
}

func (mappingGroup *MappingGroup) getResource(accessPathID, clientGroupID int64) (
	resourceID int64, err error) {

	body, err := mappingGroup.CallResourceAPI(utils.ListAPIName, nil, nil)
	if err != nil {
		return 0, errors.Annotatef(err, "list mapping groups")
	}
	records := &struct {
		MappingGroups []struct {
			ID         int64 `json:"id"`
			AccessPath struct {
				ID int64 `json:"id"`
			} `json:"access_path"`
			ClientGroup struct {
				ID int64 `json:"id"`
			} `json:"client_group"`
		} `json:"mapping_group"`
	}{}
	if err = json.Unmarshal(body, records); err != nil {
		return 0, errors.Trace(err)
	}

	for _, mappingGroupResp := range records.MappingGroups {
		if mappingGroupResp.AccessPath.ID == accessPathID &&
			mappingGroupResp.ClientGroup.ID == clientGroupID {

			return mappingGroupResp.ID, nil
		}
	}
	return 0, nil
}

// Create create the resource
func (mappingGroup *MappingGroup) Create() (created bool, err error) {
	if mappingGroup.AccessPathID == nil || mappingGroup.ClientGroupID == nil {
		return false, errors.Errorf("AccessPathID and ClientGroupID is required for resource %s",
			mappingGroup.GetType())
	}
	if config.DryRun {
		return mappingGroup.fakeCreate()
	}

	accessPathID := mappingGroup.getIntegerValue(mappingGroup.AccessPathID)
	clientGroupID := mappingGroup.getIntegerValue(mappingGroup.ClientGroupID)
	resourceID, err := mappingGroup.getResource(accessPathID, clientGroupID)
	if err != nil {
		return false, errors.Annotatef(err,
			"get mapping group with access path id %d and client group id %d",
			accessPathID, clientGroupID)
	}
	if resourceID > 0 {
		mappingGroup.repr = resourceID
		return false, nil
	}

	req := new(MappingGroupCreateReq)
	req.MappingGroup.AccessPathID = accessPathID
	req.MappingGroup.ClientGroupID = clientGroupID
	if mappingGroup.BlockVolumeIDs != nil {
		req.MappingGroup.BlockVolumeIds = mappingGroup.getIntegerListValue(mappingGroup.BlockVolumeIDs)
	}

	body, err := mappingGroup.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err,
			"create mapping group with access path id %d and client group id %d",
			accessPathID, clientGroupID)
	}
	id, status, err := mappingGroup.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	mappingGroup.repr = id
	return mappingGroup.checkStatus(status)
}
