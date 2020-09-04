package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type bucketFlag struct {
	Versioned bool `json:"versioned,omitempty"`

	Worm bool `json:"worm,omitempty"`
}

// OSBucketCreateReq defines os bucket create request
type OSBucketCreateReq struct {
	Bucket struct {

		// permission setting of all users
		AllUserPermission string `json:"all_user_permission,omitempty"`

		// permission setting of authenticated users
		AuthUserPermission string `json:"auth_user_permission,omitempty"`

		// bucket options
		Flag *bucketFlag `json:"flag,omitempty"`

		// bucket name
		Name string `json:"name"`

		// bucket owner
		OwnerID int64 `json:"owner_id"`

		// permission setting of owner
		OwnerPermission string `json:"owner_permission,omitempty"`

		// storage policy
		PolicyID int64 `json:"policy_id"`

		// max number of objects
		QuotaMaxObjects int64 `json:"quota_max_objects,omitempty"`

		// max size of all objects
		QuotaMaxSize int64 `json:"quota_max_size,omitempty"`

		// replication policy
		ReplicationPolicyID int64 `json:"replication_policy_id,omitempty"`
	} `json:"os_bucket"`
}

// ObjectStorageBucket resource
type ObjectStorageBucket struct {
	ResourceBase

	AllUserPermission  *parser.StringExpr
	AuthUserPermission *parser.StringExpr
	Name               *parser.StringExpr
	OwnerID            *parser.IntegerExpr
	OwnerPermission    *parser.StringExpr
	PolicyID           *parser.IntegerExpr
	QuotaMaxObjects    *parser.IntegerExpr
	QuotaMaxSize       *parser.IntegerExpr
}

// Init inits resource instance
func (bucket *ObjectStorageBucket) Init(stack utils.StackInterface) {
	bucket.ResourceBase.Init(stack)
	bucket.setDelegate(bucket)
}

// GetType return resource type
func (bucket *ObjectStorageBucket) GetType() string {
	return utils.ResourceObjectStorageBucket
}

// IsReady check if the formation args are ready
func (bucket *ObjectStorageBucket) IsReady() (ready bool) {
	if !bucket.isReady(bucket.AllUserPermission) ||
		!bucket.isReady(bucket.AuthUserPermission) ||
		!bucket.isReady(bucket.Name) ||
		!bucket.isReady(bucket.OwnerID) ||
		!bucket.isReady(bucket.OwnerPermission) ||
		!bucket.isReady(bucket.PolicyID) ||
		!bucket.isReady(bucket.QuotaMaxObjects) ||
		!bucket.isReady(bucket.QuotaMaxSize) {
		return false
	}

	return true
}

func (bucket *ObjectStorageBucket) fakeCreate() (bool, error) {
	bucket.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (bucket *ObjectStorageBucket) Create() (created bool, err error) {
	if bucket.Name == nil {
		return false, errors.Errorf("Name is required for resource %s", bucket.GetType())
	}
	if config.DryRun {
		return bucket.fakeCreate()
	}

	name := bucket.getStringValue(bucket.Name)
	resourceID, err := bucket.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		bucket.repr = resourceID
		return false, nil
	}

	req := new(OSBucketCreateReq)
	bucketInfo := &req.Bucket
	if bucket.AllUserPermission != nil {
		bucketInfo.AllUserPermission = bucket.getStringValue(bucket.AllUserPermission)
	}
	if bucket.AuthUserPermission != nil {
		bucketInfo.AuthUserPermission = bucket.getStringValue(bucket.AuthUserPermission)
	}
	if bucket.Name != nil {
		bucketInfo.Name = bucket.getStringValue(bucket.Name)
	}
	if bucket.OwnerID != nil {
		bucketInfo.OwnerID = bucket.getIntegerValue(bucket.OwnerID)
	}
	if bucket.OwnerPermission != nil {
		bucketInfo.OwnerPermission = bucket.getStringValue(bucket.OwnerPermission)
	}
	if bucket.PolicyID != nil {
		bucketInfo.PolicyID = bucket.getIntegerValue(bucket.PolicyID)
	}
	if bucket.QuotaMaxObjects != nil {
		bucketInfo.QuotaMaxObjects = bucket.getIntegerValue(bucket.QuotaMaxObjects)
	}
	if bucket.QuotaMaxSize != nil {
		bucketInfo.QuotaMaxSize = bucket.getIntegerValue(bucket.QuotaMaxSize)
	}

	body, err := bucket.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create bucket %s", name)
	}

	id, _, err := bucket.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	bucket.repr = id

	return false, nil
}
