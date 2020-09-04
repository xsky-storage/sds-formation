package formation

import (
	"math/rand"
	"time"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type oSKey struct {
	AccessKey string `json:"access_key,omitempty"`

	Create time.Time `json:"create,omitempty"`

	ID int64 `json:"id,omitempty"`

	Reserved bool `json:"reserved,omitempty"`

	SecretKey string `json:"secret_key,omitempty"`

	Status string `json:"status,omitempty"`

	Type string `json:"type,omitempty"`

	Update time.Time `json:"update,omitempty"`
}

// OSUserCreateReq defines os user create request
type OSUserCreateReq struct {
	OSUser struct {
		BucketQuotaMaxObjects int64 `json:"bucket_quota_max_objects,omitempty"`

		BucketQuotaMaxSize int64 `json:"bucket_quota_max_size,omitempty"`

		DisplayName string `json:"display_name,omitempty"`

		Email string `json:"email,omitempty"`

		Keys []oSKey `json:"keys,omitempty"`

		MaxBuckets int64 `json:"max_buckets,omitempty"`

		Name string `json:"name"`

		OpMask string `json:"op_mask,omitempty"`

		UserQuotaMaxObjects int64 `json:"user_quota_max_objects,omitempty"`

		UserQuotaMaxSize int64 `json:"user_quota_max_size,omitempty"`
	} `json:"os_user"`
}

// ObjectStorageKey resource
type ObjectStorageKey struct {
	AccessKey *parser.StringExpr
	SecretKey *parser.StringExpr
}

// ObjectStorageUser resource
type ObjectStorageUser struct {
	ResourceBase

	BucketQuotaMaxObjects *parser.IntegerExpr
	BucketQuotaMaxSize    *parser.IntegerExpr
	DisplayName           *parser.StringExpr
	Email                 *parser.StringExpr
	Keys                  []ObjectStorageKey
	MaxBuckets            *parser.IntegerExpr
	Name                  *parser.StringExpr
	OpMask                *parser.StringExpr
	UserQuotaMaxObjects   *parser.IntegerExpr
	UserQuotaMaxSize      *parser.IntegerExpr
}

// Init inits resource instance
func (user *ObjectStorageUser) Init(stack utils.StackInterface) {
	user.ResourceBase.Init(stack)
	user.setDelegate(user)
}

// GetType return resource type
func (user *ObjectStorageUser) GetType() string {
	return utils.ResourceObjectStorageUser
}

// IsReady check if the formation args are ready
func (user *ObjectStorageUser) IsReady() (ready bool) {
	if !user.isReady(user.BucketQuotaMaxObjects) ||
		!user.isReady(user.BucketQuotaMaxSize) ||
		!user.isReady(user.DisplayName) ||
		!user.isReady(user.Email) ||
		!user.isReady(user.MaxBuckets) ||
		!user.isReady(user.Name) ||
		!user.isReady(user.OpMask) ||
		!user.isReady(user.UserQuotaMaxObjects) ||
		!user.isReady(user.UserQuotaMaxSize) {
		return false
	}
	for _, key := range user.Keys {
		if !user.isReady(key.AccessKey) ||
			!user.isReady(key.SecretKey) {
			return false
		}
	}

	return true
}

func (user *ObjectStorageUser) fakeCreate() (bool, error) {
	user.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (user *ObjectStorageUser) Create() (created bool, err error) {
	if user.Name == nil {
		err = errors.Errorf("Name is required for resource %s", user.GetType())
		return
	}
	if config.DryRun {
		return user.fakeCreate()
	}

	name := user.getStringValue(user.Name)
	resourceID, err := user.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		user.repr = resourceID
		return false, nil
	}

	req := new(OSUserCreateReq)
	userInfo := &req.OSUser
	if user.BucketQuotaMaxObjects != nil {
		userInfo.BucketQuotaMaxObjects = user.getIntegerValue(user.BucketQuotaMaxObjects)
	}
	if user.BucketQuotaMaxSize != nil {
		userInfo.BucketQuotaMaxSize = user.getIntegerValue(user.BucketQuotaMaxSize)
	}
	if user.DisplayName != nil {
		userInfo.DisplayName = user.getStringValue(user.DisplayName)
	}
	if user.Email != nil {
		userInfo.Email = user.getStringValue(user.Email)
	}
	if user.MaxBuckets != nil {
		userInfo.MaxBuckets = user.getIntegerValue(user.MaxBuckets)
	}
	if user.Name != nil {
		userInfo.Name = user.getStringValue(user.Name)
	}
	if user.OpMask != nil {
		userInfo.OpMask = user.getStringValue(user.OpMask)
	}
	if user.UserQuotaMaxObjects != nil {
		userInfo.UserQuotaMaxObjects = user.getIntegerValue(user.UserQuotaMaxObjects)
	}
	if user.UserQuotaMaxSize != nil {
		userInfo.UserQuotaMaxSize = user.getIntegerValue(user.UserQuotaMaxSize)
	}
	for _, userKey := range user.Keys {
		key := &oSKey{
			AccessKey: user.getStringValue(userKey.AccessKey),
			SecretKey: user.getStringValue(userKey.SecretKey),
		}
		key.AccessKey = user.getStringValue(userKey.AccessKey)
		key.SecretKey = user.getStringValue(userKey.SecretKey)
		userInfo.Keys = append(userInfo.Keys, *key)
	}

	body, err := user.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create object storage user %s", name)
	}

	id, status, err := user.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	user.repr = id
	created, err = user.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
