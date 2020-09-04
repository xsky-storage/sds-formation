package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSActiveDirectoryCreateReq defines request to create file storage active directory
type FSActiveDirectoryCreateReq struct {
	Info struct {
		// name of active directory
		Name string `json:"name" required:"true"`
		// workgroup of active directory
		Workgroup string `json:"workgroup,omitempty" required:"true"`
		// realm of active directory
		Realm string `json:"realm,omitempty" required:"true"`
		// ip of dns server
		IP string `json:"ip,omitempty" required:"false"`
		// username of active directory
		Username string `json:"username,omitempty" required:"true"`
		// password of active directory
		Password string `json:"password,omitempty" required:"true"`
	} `json:"fs_active_directory" required:"true"`
}

// FSAD resource
type FSAD struct {
	ResourceBase

	Name      *parser.StringExpr
	Workgroup *parser.StringExpr
	Realm     *parser.StringExpr
	IP        *parser.StringExpr
	UserName  *parser.StringExpr
	Password  *parser.StringExpr
}

// Init inits resource instance
func (ad *FSAD) Init(stack utils.StackInterface) {
	ad.ResourceBase.Init(stack)
	ad.setDelegate(ad)
}

// GetType return resource type
func (ad *FSAD) GetType() string {
	return utils.ResourceFSAD
}

// IsReady check if the formation args are ready
func (ad *FSAD) IsReady() (ready bool) {
	if !ad.isReady(ad.Name) ||
		!ad.isReady(ad.Workgroup) ||
		!ad.isReady(ad.Realm) ||
		!ad.isReady(ad.IP) ||
		!ad.isReady(ad.UserName) ||
		!ad.isReady(ad.Password) {
		return false
	}

	return true
}

func (ad *FSAD) fakeCreate() (bool, error) {
	ad.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (ad *FSAD) Create() (created bool, err error) {
	if ad.Name == nil {
		err = errors.Errorf("Name is required for resource %s", ad.GetType())
		return
	}
	if config.DryRun {
		return ad.fakeCreate()
	}

	name := ad.getStringValue(ad.Name)
	resourceID, err := ad.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		ad.repr = resourceID
		return false, nil
	}

	req := new(FSActiveDirectoryCreateReq)
	userInfo := &req.Info
	if ad.Name != nil {
		userInfo.Name = ad.getStringValue(ad.Name)
	}
	if ad.Workgroup != nil {
		userInfo.Workgroup = ad.getStringValue(ad.Workgroup)
	}
	if ad.Realm != nil {
		userInfo.Realm = ad.getStringValue(ad.Realm)
	}
	if ad.IP != nil {
		userInfo.IP = ad.getStringValue(ad.IP)
	}
	if ad.UserName != nil {
		userInfo.Username = ad.getStringValue(ad.UserName)
	}
	if ad.Password != nil {
		userInfo.Password = ad.getStringValue(ad.Password)
	}

	body, err := ad.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs active directory %s", name)
	}

	id, status, err := ad.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	ad.repr = id
	created, err = ad.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return
}
