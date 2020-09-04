package formation

import (
	"encoding/json"
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// NetworkAddress resource
type NetworkAddress struct {
	ResourceBase

	IP *parser.StringExpr
}

// Init inits resource instance
func (address *NetworkAddress) Init(stack utils.StackInterface) {
	address.ResourceBase.Init(stack)
	address.setDelegate(address)
}

// GetType return resource type
func (address *NetworkAddress) GetType() string {
	return utils.ResourceNetworkAddress
}

// IsReady check if the formation args are ready
func (address *NetworkAddress) IsReady() (ready bool) {
	if !address.isReady(address.IP) {
		return false
	}

	return true
}

// Get get resource from server
func (address *NetworkAddress) Get() error {
	if address.IP == nil {
		return errors.Errorf("IP is needed for get netword address")
	}
	if config.DryRun {
		address.repr = rand.Int63()
		return nil
	}

	// no limit when list resource
	queryParam := map[string]string{"limit": "-1"}
	body, err := address.CallResourceAPI(utils.ListAPIName, nil, nil, queryParam)
	if err != nil {
		return errors.Trace(err)
	}
	addressRecords := map[string]json.RawMessage{}
	if err = json.Unmarshal(body, &addressRecords); err != nil {
		return errors.Trace(err)
	}
	recordsKey, err := settings.GetSetting(utils.ResourceNetworkAddress, utils.RecordsKey)
	if err != nil {
		return errors.Trace(err)
	}
	addressesData, ok := addressRecords[recordsKey]
	if !ok {
		return errors.Errorf("key %s not found in list address response", recordsKey)
	}

	addressesResp := []*struct {
		ID int64  `json:"id"`
		IP string `json:"ip"`
	}{}
	if err = json.Unmarshal(addressesData, &addressesResp); err != nil {
		return errors.Trace(err)
	}
	var found bool
	ip := address.getStringValue(address.IP)
	for _, addr := range addressesResp {
		if addr.IP == ip {
			found = true
			address.repr = addr.ID
			break
		}
	}

	if !found {
		return errors.Errorf("no network address found with %s", ip)
	}

	return nil
}
