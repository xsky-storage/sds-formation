package formation

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// OSPolicyCreateReq defines os policy create request
type OSPolicyCreateReq struct {
	Policy struct {
		Name                string                 `json:"name"`
		Description         string                 `json:"description,omitempty"`
		IndexPoolID         int64                  `json:"index_pool_id"`
		CachePoolID         int64                  `json:"cache_pool_id,omitempty"`
		DataPoolID          int64                  `json:"data_pool_id,omitempty"`
		DataPoolIDs         []int64                `json:"data_pool_ids,omitempty"`
		Shared              *bool                  `json:"shared,omitempty"`
		Compress            *bool                  `json:"compress,omitempty"`
		Crypto              *bool                  `json:"crypto,omitempty"`
		ObjectSizeThreshold int64                  `json:"object_size_threshold,omitempty"`
		StorageClasses      []*oSStorageClassInReq `json:"storage_classes,omitempty"`
	} `json:"os_policy"`
}

type oSStorageClassInReq struct {
	Name            string  `json:"name"`
	Class           int     `json:"class"`
	Description     string  `json:"description,omitempty"`
	ActivePoolIDs   []int64 `json:"active_pool_ids"`
	InactivePoolIDs []int64 `json:"inactive_pool_ids,omitempty"`
}

// OSStorageClass resource
type OSStorageClass struct {
	Name            *parser.StringExpr
	ClassID         *parser.StringExpr
	Description     *parser.StringExpr
	ActivePoolIDs   *parser.IntegerListExpr
	InactivePoolIDs *parser.IntegerListExpr
}

// ObjectStoragePolicy resource
type ObjectStoragePolicy struct {
	ResourceBase

	PolicyReqVersion    *parser.IntegerExpr
	Compress            *parser.BoolExpr
	Crypto              *parser.BoolExpr
	DataPoolID          *parser.IntegerExpr
	DataPoolIDs         *parser.IntegerListExpr
	Description         *parser.StringExpr
	IndexPoolID         *parser.IntegerExpr
	CachePoolID         *parser.IntegerExpr
	Shared              *parser.BoolExpr
	Name                *parser.StringExpr
	ObjectSizeThreshold *parser.IntegerExpr
	StorageClasses      []*OSStorageClass
}

// Init inits resource instance
func (policy *ObjectStoragePolicy) Init(stack utils.StackInterface) {
	policy.ResourceBase.Init(stack)
	policy.setDelegate(policy)
}

// GetType return resource type
func (policy *ObjectStoragePolicy) GetType() string {
	return utils.ResourceObjectStoragePolicy
}

// IsReady check if the formation args are ready
func (policy *ObjectStoragePolicy) IsReady() (ready bool) {
	if !policy.isReady(policy.Compress) ||
		!policy.isReady(policy.Crypto) ||
		!policy.isReady(policy.DataPoolID) ||
		!policy.isReady(policy.DataPoolIDs) ||
		!policy.isReady(policy.Description) ||
		!policy.isReady(policy.IndexPoolID) ||
		!policy.isReady(policy.CachePoolID) ||
		!policy.isReady(policy.Shared) ||
		!policy.isReady(policy.Name) ||
		!policy.isReady(policy.ObjectSizeThreshold) {

		return false
	}

	for _, sc := range policy.StorageClasses {
		if !policy.isReady(sc.Name) ||
			!policy.isReady(sc.ClassID) ||
			!policy.isReady(sc.Description) ||
			!policy.isReady(sc.ActivePoolIDs) ||
			!policy.isReady(sc.InactivePoolIDs) {

			return false
		}
	}

	return true
}

func (policy *ObjectStoragePolicy) fakeCreate() (bool, error) {
	policy.repr = rand.Int63()
	return true, nil
}

func (policy *ObjectStoragePolicy) parseReq41LaterFormatTemplate() *OSPolicyCreateReq {
	req := new(OSPolicyCreateReq)

	for _, scData := range policy.StorageClasses {
		sc := &oSStorageClassInReq{
			Name:          policy.getStringValue(scData.Name),
			Class:         0, // only class_0 supported currently
			ActivePoolIDs: policy.getIntegerListValue(scData.ActivePoolIDs),
		}
		if scData.Description != nil {
			sc.Description = policy.getStringValue(scData.Description)
		}
		if scData.InactivePoolIDs != nil {
			sc.InactivePoolIDs = policy.getIntegerListValue(scData.InactivePoolIDs)
		}
		req.Policy.StorageClasses = append(req.Policy.StorageClasses, sc)
		break
	}

	return req
}

func (policy *ObjectStoragePolicy) parseReqFromOldFormatTemplate() *OSPolicyCreateReq {

	req := new(OSPolicyCreateReq)
	if policy.DataPoolID != nil {
		sc := new(oSStorageClassInReq)
		sc.Name = "default"
		sc.Class = 0
		sc.ActivePoolIDs = []int64{policy.getIntegerValue(policy.DataPoolID)}
		if policy.DataPoolIDs != nil {
			activePoolMap := map[int64]bool{}
			for _, activePoolID := range sc.ActivePoolIDs {
				activePoolMap[activePoolID] = true
			}
			for _, inactivePoolID := range policy.getIntegerListValue(policy.DataPoolIDs) {
				if !activePoolMap[inactivePoolID] {
					sc.InactivePoolIDs = append(sc.InactivePoolIDs, inactivePoolID)
				}
			}
		}
		req.Policy.StorageClasses = append(req.Policy.StorageClasses, sc)
	}
	return req
}

// Create create the resource
func (policy *ObjectStoragePolicy) Create() (created bool, err error) {
	if policy.Name == nil {
		return false, fmt.Errorf("Name is required for resource %s", policy.GetType())
	}
	if config.DryRun {
		return policy.fakeCreate()
	}

	name := policy.getStringValue(policy.Name)
	resourceID, err := policy.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		policy.repr = resourceID
		return true, nil
	}

	reqVersion := int64(2)
	serverVersion := policy.stack.GetOpenAPIClient().ServerVersion()
	if !strings.HasPrefix(`SDS_`, serverVersion) {
		reqVersion = 1
	}
	if policy.PolicyReqVersion != nil {
		reqVersion = policy.getIntegerValue(policy.PolicyReqVersion)
	}

	var req *OSPolicyCreateReq
	if reqVersion > 1 {
		if policy.StorageClasses != nil {
			req = policy.parseReq41LaterFormatTemplate()
		} else {
			req = policy.parseReqFromOldFormatTemplate()
		}
	} else {
		req = new(OSPolicyCreateReq)
		if policy.DataPoolID != nil {
			req.Policy.DataPoolID = policy.getIntegerValue(policy.DataPoolID)
		}
		if policy.DataPoolIDs != nil {
			req.Policy.DataPoolIDs = policy.getIntegerListValue(policy.DataPoolIDs)
		}
	}

	if policy.Compress != nil {
		compress := policy.getBoolValue(policy.Compress)
		req.Policy.Compress = &compress
	}
	if policy.Crypto != nil {
		crypto := policy.getBoolValue(policy.Crypto)
		req.Policy.Crypto = &crypto
	}
	if policy.Description != nil {
		req.Policy.Description = policy.getStringValue(policy.Description)
	}
	if policy.IndexPoolID != nil {
		req.Policy.IndexPoolID = policy.getIntegerValue(policy.IndexPoolID)
	}
	if policy.Name != nil {
		req.Policy.Name = policy.getStringValue(policy.Name)
	}
	if policy.ObjectSizeThreshold != nil {
		req.Policy.ObjectSizeThreshold = policy.getIntegerValue(policy.ObjectSizeThreshold)
	}

	if policy.CachePoolID != nil {
		req.Policy.CachePoolID = policy.getIntegerValue(policy.CachePoolID)
	}
	if policy.Shared != nil {
		shared := policy.getBoolValue(policy.Shared)
		req.Policy.Shared = &shared
	}

	body, err := policy.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create object storage policy %s", name)
	}
	id, _, err := policy.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	policy.repr = id
	return false, nil
}
