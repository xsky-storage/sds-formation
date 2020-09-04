package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

// FSLdapCreateReq defines request to create file storage ldap
type FSLdapCreateReq struct {
	Info struct {
		// name of ldap
		Name string `json:"name" required:"true"`
		// ip of server
		IP string `json:"ip" required:"true"`
		// ips of standby servers
		IPs []string `json:"ips,omitempty"`
		// ldap service port
		Port int `json:"port,omitempty" required:"true"`
		// ldap suffix
		Suffix string `json:"suffix,omitempty" required:"true"`
		// ldap admin dn
		AdminDN string `json:"admin_dn,omitempty"`
		// bind password
		Password string `json:"password,omitempty"`
		// ldap suffix
		UserSuffix string `json:"user_suffix,omitempty"`
		// group suffix
		GroupSuffix string `json:"group_suffix,omitempty"`
		// timeout for searching
		Timeout int64 `json:"timeout,omitempty"`
		// timeout for connection
		ConnectionTimeout int64 `json:"connection_timeout,omitempty"`
	} `json:"fs_ldap" required:"true"`
}

// FSLdap resource
type FSLdap struct {
	ResourceBase

	Name              *parser.StringExpr
	IP                *parser.StringExpr
	IPs               *parser.StringListExpr
	Port              *parser.IntegerExpr
	Suffix            *parser.StringExpr
	AdminDN           *parser.StringExpr
	Password          *parser.StringExpr
	UserSuffix        *parser.StringExpr
	GroupSuffix       *parser.StringExpr
	Timeout           *parser.IntegerExpr
	ConnectionTimeout *parser.IntegerExpr
}

// Init inits resource instance
func (ldap *FSLdap) Init(stack utils.StackInterface) {
	ldap.ResourceBase.Init(stack)
	ldap.setDelegate(ldap)
}

// GetType return resource type
func (ldap *FSLdap) GetType() string {
	return utils.ResourceFSLdap
}

// IsReady check if the formation args are ready
func (ldap *FSLdap) IsReady() (ready bool) {
	if !ldap.isReady(ldap.Name) ||
		!ldap.isReady(ldap.IP) ||
		!ldap.isReady(ldap.IPs) ||
		!ldap.isReady(ldap.Port) ||
		!ldap.isReady(ldap.Suffix) ||
		!ldap.isReady(ldap.AdminDN) ||
		!ldap.isReady(ldap.Password) ||
		!ldap.isReady(ldap.UserSuffix) ||
		!ldap.isReady(ldap.GroupSuffix) ||
		!ldap.isReady(ldap.ConnectionTimeout) ||
		!ldap.isReady(ldap.Timeout) {
		return false
	}
	return true
}

func (ldap *FSLdap) fakeCreate() (bool, error) {
	ldap.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (ldap *FSLdap) Create() (created bool, err error) {
	if ldap.Name == nil {
		err = errors.Errorf("Name is required for resource %s", ldap.GetType())
		return
	}
	if config.DryRun {
		return ldap.fakeCreate()
	}

	name := ldap.getStringValue(ldap.Name)
	resourceID, err := ldap.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		ldap.repr = resourceID
		return false, nil
	}

	req := new(FSLdapCreateReq)
	userInfo := &req.Info
	if ldap.Name != nil {
		userInfo.Name = ldap.getStringValue(ldap.Name)
	}
	if ldap.IP != nil {
		userInfo.IP = ldap.getStringValue(ldap.IP)
	}
	if ldap.IPs != nil {
		userInfo.IPs = ldap.getStringListValue(ldap.IPs)
	}
	if ldap.Port != nil {
		userInfo.Port = int(ldap.getIntegerValue(ldap.Port))
	}
	if ldap.Suffix != nil {
		userInfo.Suffix = ldap.getStringValue(ldap.Suffix)
	}
	if ldap.AdminDN != nil {
		userInfo.AdminDN = ldap.getStringValue(ldap.AdminDN)
	}
	if ldap.Password != nil {
		userInfo.Password = ldap.getStringValue(ldap.Password)
	}
	if ldap.UserSuffix != nil {
		userInfo.UserSuffix = ldap.getStringValue(ldap.UserSuffix)
	}
	if ldap.GroupSuffix != nil {
		userInfo.GroupSuffix = ldap.getStringValue(ldap.GroupSuffix)
	}
	if ldap.Timeout != nil {
		userInfo.Timeout = ldap.getIntegerValue(ldap.Timeout)
	}
	if ldap.ConnectionTimeout != nil {
		userInfo.ConnectionTimeout = ldap.getIntegerValue(ldap.ConnectionTimeout)
	}

	body, err := ldap.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create fs ldap %s", name)
	}

	id, status, err := ldap.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}

	ldap.repr = id
	created, err = ldap.checkStatus(status)
	if err != nil {
		return false, errors.Trace(err)
	}
	return created, nil
}
