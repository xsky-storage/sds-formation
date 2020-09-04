package formation

import (
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type templateTestSuite struct {
	suite.Suite
}

func (s *templateTestSuite) TestLoadTemplate() {
	template := `{
	"Description": "this will",
	"Parameters": {
		"ClusterURL": {
		    "Type": "String",
		    "Value": "http://10.0.0.1:8056/v1"
		},
		"admin_ip": {
			"Type": "String",
			"Value": "10.0.0.2"
		},
		"name": {
			"Value": "admin",
			"Type": "String"
		},
		"pwd": {
			"Value": "admin",
			"Type": "String"
		}
	},
	"Resources": [
		{
			"Name": "token",
			"Type": "Token",
			"Properties": {
				"Name": {"Ref": "name"},
				"Password": {"Ref": "pwd"}
			}
		},
		{
			"Name": "disk_list",
			"Type": "DiskList",
			"Action": "Get",
			"Properties": {
				"Used": false,
				"DiskType": "SSD"
			}
		},
		{
			"Name": "osd",
			"Type": "Osd",
			"Action": "Create",
			"Properties": {
				"DiskID": {"Select": [0, {"Ref": "disk_list"}]}
			}
		},
		{
			"Name": "pool",
			"Type": "Pool",
			"Action": "Create",
			"Properties": {
				"Name": "pool1",
				"Size": 1,
				"PoolType": "replicated",
				"OsdIDs": [{"Ref": "osd"}]
			}
		},
		{
			"Name": "policy1",
			"Type": "ObjectStoragePolicy",
			"Properties": {
				"Name": "policyccqc",
				"Compress": true,
				"Crypto": true,
				"Shared": false,
				"IndexPoolID": 1,
				"CachePoolID": 1,
				"DataPoolID": {"Ref": "pool"}
			}
		}
	]
}`
	tpl := new(Template)

	assert.NoError(s.T(), json.Unmarshal([]byte(template), tpl))

}
