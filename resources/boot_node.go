package formation

import (
	"encoding/json"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// BootNodeReq defines boot node initialize request
type BootNodeReq struct {
	BootNode struct {

		// admin network: 10.0.0.0/24
		AdminNetwork string `json:"admin_network,omitempty"`

		// gateway networks, multiple networks splited by comma: 10.0.3.0/24,10.0.4.0/24
		GatewayNetwork string `json:"gateway_network,omitempty"`

		// path of sds installer packages
		InstallerPath string `json:"installer_path,omitempty"`

		// private network : 10.0.2.0/24
		PrivateNetwork string `json:"private_network"`

		// public network: 10.0.1.0/24
		PublicNetwork string `json:"public_network"`
	} `json:"bootnode"`
}

// BootNodeResp defines response of boot node
type BootNodeResp struct {
	BootNode struct {
		Host struct {
		}
	} `json:"boot_node"`
}

// BootNode resource
type BootNode struct {
	ResourceBase
	AdminNetwork   *parser.StringExpr
	PrivateNetwork *parser.StringExpr
	PublicNetwork  *parser.StringExpr
	GatewayNetwork *parser.StringExpr
	InstallerPath  *parser.StringExpr
}

// Init inits resource instance
func (bootNode *BootNode) Init(stack utils.StackInterface) {
	bootNode.ResourceBase.Init(stack)
	bootNode.setDelegate(bootNode)
}

// GetType return resource type
func (bootNode *BootNode) GetType() string {
	return utils.ResourceBootNode
}

// IsReady check if the formation args are ready
func (bootNode *BootNode) IsReady() (ready bool) {
	if !bootNode.isReady(bootNode.AdminNetwork) ||
		!bootNode.isReady(bootNode.PrivateNetwork) ||
		!bootNode.isReady(bootNode.PublicNetwork) ||
		!bootNode.isReady(bootNode.GatewayNetwork) ||
		!bootNode.isReady(bootNode.InstallerPath) {
		return false
	}

	return true
}

func (bootNode *BootNode) fakeCreate() (bool, error) {
	bootNode.repr = int64(1)
	return true, nil
}

// Create create the resource
func (bootNode *BootNode) Create() (created bool, err error) {
	if config.DryRun {
		return bootNode.fakeCreate()
	}

	req := new(BootNodeReq)
	bootNodeInfo := &req.BootNode
	if bootNode.AdminNetwork != nil {
		bootNodeInfo.AdminNetwork = bootNode.getStringValue(bootNode.AdminNetwork)
	}
	if bootNode.PrivateNetwork != nil {
		bootNodeInfo.PrivateNetwork = bootNode.getStringValue(bootNode.PrivateNetwork)
	}
	if bootNode.PublicNetwork != nil {
		bootNodeInfo.PublicNetwork = bootNode.getStringValue(bootNode.PublicNetwork)
	}
	if bootNode.GatewayNetwork != nil {
		bootNodeInfo.GatewayNetwork = bootNode.getStringValue(bootNode.GatewayNetwork)
	}
	if bootNode.InstallerPath != nil {
		bootNodeInfo.InstallerPath = bootNode.getStringValue(bootNode.InstallerPath)
	}

	body, err := bootNode.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "failed to set boot node")
	}

	record := &struct {
		BootNode struct {
			Host struct {
				ID int64 `json:"id"`
			} `json:"host"`
		} `json:"bootnode"`
	}{}
	if err = json.Unmarshal(body, record); err != nil {
		return false, errors.Trace(err)
	}
	// NOTE: Save host id of boot node as the representation of boot node
	bootNode.repr = record.BootNode.Host.ID

	return false, nil
}
