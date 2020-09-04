package formation

import (
	"math/rand"

	"github.com/juju/errors"

	"xsky.com/sds-formation/config"
	"xsky.com/sds-formation/parser"
	"xsky.com/sds-formation/utils"
)

type s3LoadBalancerGroupCreateReqGroupLoadBalancersElt struct {

	// host of load balancer
	HostID int64 `json:"host_id"`

	// vip will be bound to interface, exclusive to ip
	InterfaceName string `json:"interface_name,omitempty"`

	// vip will be bound to interface of the gateway ip, exclusive to interface_name
	IP string `json:"ip,omitempty"`

	// virtual ip of load balancer
	Vip string `json:"vip"`
}

// S3LBGroupCreateReq defines s3 lb group create request
type S3LBGroupCreateReq struct {
	S3LoadBalancerGroupCreateReqGroup struct {

		// group description
		Description string `json:"description,omitempty"`

		// group port
		HTTPSPort int64 `json:"https_port,omitempty"`

		// group name
		Name string `json:"name"`

		// group port
		Port int64 `json:"port"`

		// s3 load balancers
		S3LoadBalancers []s3LoadBalancerGroupCreateReqGroupLoadBalancersElt `json:"s3_load_balancers"`
	} `json:"s3_load_balancer_group"`
}

// S3LoadBalancer resource
type S3LoadBalancer struct {
	HostID *parser.IntegerExpr
	VIP    *parser.StringExpr
}

// S3LoadBalancerGroup resource
type S3LoadBalancerGroup struct {
	ResourceBase
	Name            *parser.StringExpr
	Port            *parser.IntegerExpr
	S3LoadBalancers []*S3LoadBalancer
}

// Init inits resource instance
func (lbg *S3LoadBalancerGroup) Init(stack utils.StackInterface) {
	lbg.ResourceBase.Init(stack)
	lbg.setDelegate(lbg)
}

// GetType return resource type
func (lbg *S3LoadBalancerGroup) GetType() string {
	return utils.ResourceS3LoadBalancerGroup
}

// IsReady check if the formation args are ready
func (lbg *S3LoadBalancerGroup) IsReady() bool {
	if !lbg.isReady(lbg.Name) ||
		!lbg.isReady(lbg.Port) {
		return false
	}
	for _, lb := range lbg.S3LoadBalancers {
		if !lbg.isReady(lb.HostID) ||
			!lbg.isReady(lb.VIP) {
			return false
		}
	}
	return true
}

func (lbg *S3LoadBalancerGroup) fakeCreate() (bool, error) {
	lbg.repr = rand.Int63()
	return true, nil
}

// Create create the resource
func (lbg *S3LoadBalancerGroup) Create() (created bool, err error) {
	if lbg.Name == nil {
		return false, errors.Errorf("Name is required for resource %s", lbg.GetType())
	}
	if config.DryRun {
		return lbg.fakeCreate()
	}

	name := lbg.getStringValue(lbg.Name)
	resourceID, err := lbg.getResourceByName(name)
	if err != nil {
		return false, errors.Trace(err)
	}
	if resourceID != nil {
		lbg.repr = resourceID
		return false, nil
	}

	req := new(S3LBGroupCreateReq)
	lbGroupInfo := &req.S3LoadBalancerGroupCreateReqGroup
	lbGroupInfo.Name = name
	lbGroupInfo.Port = lbg.getIntegerValue(lbg.Port)
	for _, lb := range lbg.S3LoadBalancers {
		balancerReq := s3LoadBalancerGroupCreateReqGroupLoadBalancersElt{
			HostID: lbg.getIntegerValue(lb.HostID),
			Vip:    lbg.getStringValue(lb.VIP),
		}
		lbGroupInfo.S3LoadBalancers = append(lbGroupInfo.S3LoadBalancers, balancerReq)
	}

	body, err := lbg.CallCreateAPI(req, nil)
	if err != nil {
		return false, errors.Annotatef(err, "create s3 load balancer group %s", name)
	}

	id, _, err := lbg.getIdentifyAndStatus(body)
	if err != nil {
		return false, errors.Trace(err)
	}
	lbg.repr = id

	return false, nil
}
