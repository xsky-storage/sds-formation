package formation

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// Hosts resource
type Hosts struct {
	ResourceBase
	AdminIPs           *parser.StringListExpr
	ProtectionDomainID *parser.IntegerExpr
	Description        *parser.StringExpr
	Roles              *parser.StringListExpr
	Type               *parser.StringExpr
	creatingHostIDs    []int64
}

// Init inits resource instance
func (hosts *Hosts) Init(stack utils.StackInterface) {
	hosts.ResourceBase.Init(stack)
	hosts.setDelegate(hosts)
}

// GetType return resource type
func (hosts *Hosts) GetType() string {
	return utils.ResourceHosts
}

// CheckInterval return check interval
func (hosts *Hosts) CheckInterval() int {
	return 15 * len(hosts.creatingHostIDs)
}

// IsReady check if the formation args are ready
func (hosts *Hosts) IsReady() (ready bool) {
	if !hosts.isReady(hosts.AdminIPs) ||
		!hosts.isReady(hosts.ProtectionDomainID) ||
		!hosts.isReady(hosts.Description) ||
		!hosts.isReady(hosts.Roles) {
		return false
	}

	return true
}

func (hosts *Hosts) fakeCreate() (bool, error) {
	hostIDs := []int64{}
	adminIPs := hosts.getStringListValue(hosts.AdminIPs)
	for range adminIPs {
		hostIDs = append(hostIDs, rand.Int63())
	}
	hosts.repr = hostIDs
	return true, nil
}

// Get get resource from server
func (hosts *Hosts) Get() error {
	if config.DryRun {
		hosts.repr = []int64{rand.Int63(), rand.Int63()}
		return nil
	}

	// no limit when list resource
	queryParam := map[string]string{"limit": "-1"}
	body, err := hosts.CallResourceAPI(utils.ListAPIName, nil, nil, queryParam)
	if err != nil {
		return errors.Trace(err)
	}
	hostRecords := map[string]json.RawMessage{}
	if err = json.Unmarshal(body, &hostRecords); err != nil {
		return errors.Trace(err)
	}
	recordsKey, err := settings.GetSetting(utils.ResourceHosts, utils.RecordsKey)
	if err != nil {
		return errors.Trace(err)
	}
	hostsData, ok := hostRecords[recordsKey]
	if !ok {
		return errors.Errorf("key %s not found in list hosts response", recordsKey)
	}

	hostsResp := []*struct {
		ID    int64  `json:"id"`
		Roles string `json:"roles"`
		Type  string `json:"type"`
	}{}
	if err = json.Unmarshal(hostsData, &hostsResp); err != nil {
		return errors.Trace(err)
	}

	var roles []string
	if hosts.Roles != nil {
		roles = hosts.getStringListValue(hosts.Roles)
	}
	hostType := ""
	if hosts.Type != nil {
		hostType = hosts.getStringValue(hosts.Type)
	}
	ids := []int64{}
	for _, host := range hostsResp {
		if host.Type != hostType && hostType != "" {
			continue
		}
		if len(roles) != 0 {
			for _, role := range roles {
				if strings.Index(host.Roles, role) != -1 {
					ids = append(ids, host.ID)
					break
				}
			}
		} else {
			ids = append(ids, host.ID)
		}
	}
	hosts.repr = ids
	return nil
}

// Create create the resource
func (hosts *Hosts) Create() (created bool, err error) {
	if hosts.AdminIPs == nil {
		err = errors.Errorf("AdminIPs is required")
		return
	}

	if config.DryRun {
		return hosts.fakeCreate()
	}

	req := new(HostCreateReq)
	if hosts.ProtectionDomainID != nil {
		req.Host.ProtectionDomainID = hosts.getIntegerValue(hosts.ProtectionDomainID)
	}
	if hosts.Description != nil {
		req.Host.Description = hosts.getStringValue(hosts.Description)
	}
	if hosts.Roles != nil {
		req.Host.Roles = hosts.getStringListValue(hosts.Roles)
	}
	if hosts.Type != nil {
		req.Host.Type = hosts.getStringValue(hosts.Type)
	}

	hostIDs := []int64{}
	adminIPs := hosts.getStringListValue(hosts.AdminIPs)
	for _, adminIP := range adminIPs {
		req.Host.AdminIP = adminIP
		body, err := hosts.CallCreateAPI(req, nil)
		if err != nil {
			return false, errors.Annotatef(err, "create host with admin ip %s", adminIP)
		}
		id, _, err := hosts.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}
		hostIDs = append(hostIDs, id.(int64))
	}

	hosts.repr = hostIDs
	hosts.creatingHostIDs = hostIDs
	return false, nil
}

// IsCreated if the resource has been created
func (hosts *Hosts) IsCreated() (created bool, err error) {
	creatingHostIDs := []int64{}
	getReqIdentifyKey, err := settings.GetSetting(hosts.GetType(), utils.GetReqIdentify)
	if err != nil {
		return false, errors.Trace(err)
	}
	for _, hostID := range hosts.creatingHostIDs {
		pathParam := map[string]string{getReqIdentifyKey: fmt.Sprintf("%d", hostID)}
		body, err := hosts.CallGetAPI(pathParam)
		if err != nil {
			return false, errors.Annotatef(err, "get host with id %d", hostID)
		}
		_, status, err := hosts.getIdentifyAndStatus(body)
		if err != nil {
			return false, errors.Trace(err)
		}
		created, err = hosts.checkStatus(status)
		if err != nil {
			return false, errors.Trace(err)
		}
		if created {
			log.Printf("item host %d is created", hostID)
		} else {
			creatingHostIDs = append(creatingHostIDs, hostID)
		}
	}

	if len(creatingHostIDs) == 0 {
		return true, nil
	}

	hosts.creatingHostIDs = creatingHostIDs
	return false, nil
}
