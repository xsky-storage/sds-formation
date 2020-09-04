package formation

import (
	"github.com/juju/errors"

	"xsky.com/sds-formation/utils"
)

type resourceSettings map[string]map[string]string

// GetSetting returns resource setting
func (s resourceSettings) GetSetting(resource, key string) (string, error) {
	r, ok := s[resource]
	if !ok {
		return "", errors.Errorf("unsupport resource %s", resource)
	}
	val, ok := r[key]
	if !ok {
		return "", errors.Errorf("value of %s's %s not found", resource, key)
	}
	return val, nil
}

// GetStatusKey returns resource status key
func (s resourceSettings) GetStatusKey(resource string) string {
	val, err := s.GetSetting(resource, utils.StatusKey)
	if err != nil {
		return "Status"
	}
	return val
}

// GetIdentifyKey returns resource identify key
func (s resourceSettings) GetIdentifyKey(resource string) string {
	val, err := s.GetSetting(resource, utils.IdentifyKey)
	if err != nil {
		return "ID"
	}
	return val
}

var settings = resourceSettings{
	utils.ResourceAccessPath: {
		utils.GetReqIdentify: "access_path_id",
		utils.ListAPIName:    "ListAccessPaths",
		utils.GetAPIName:     "GetAccessPath",
		utils.RecordKey:      "access_path",
		utils.RecordsKey:     "access_paths",
		utils.CreateAPIName:  "CreateAccessPath",
	},
	utils.ResourceBlockVolume: {
		utils.GetReqIdentify: "block_volume_id",
		utils.GetAPIName:     "GetBlockVolume",
		utils.ListAPIName:    "ListBlockVolumes",
		utils.RecordsKey:     "block_volumes",
		utils.RecordKey:      "block_volume",
		utils.CreateAPIName:  "CreateBlockVolume",
	},
	utils.ResourceBlockVolumes: {
		utils.ListAPIName:    "ListBlockVolumes",
		utils.GetReqIdentify: "block_volume_id",
		utils.GetAPIName:     "GetBlockVolume",
		utils.RecordKey:      "block_volume",
		utils.RecordsKey:     "block_volumes",
		utils.CreateAPIName:  "CreateBlockVolume",
	},
	utils.ResourceBootNode: {
		utils.GetAPIName:    "BootNode",
		utils.RecordKey:     "bootnode",
		utils.CreateAPIName: "SetBootNode",
	},
	utils.ResourceClientGroup: {
		utils.GetReqIdentify: "client_group_id",
		utils.ListAPIName:    "ListClientGroups",
		utils.GetAPIName:     "GetClientGroup",
		utils.RecordKey:      "client_group",
		utils.RecordsKey:     "client_groups",
		utils.CreateAPIName:  "CreateClientGroup",
	},
	utils.ResourceDiskList: {
		utils.ListAPIName:    "ListDisks",
		utils.UpdateAPIName:  "UpdateDisk",
		utils.GetReqIdentify: "disk_id",
	},
	utils.ResourceHost: {
		utils.RecordKey:      "host",
		utils.GetAPIName:     "GetHost",
		utils.GetReqIdentify: "host_id",
		utils.CreateAPIName:  "CreateHost",
	},
	utils.ResourceHosts: {
		utils.RecordsKey:     "hosts",
		utils.ListAPIName:    "ListHosts",
		utils.RecordKey:      "host",
		utils.CreateAPIName:  "CreateHost",
		utils.GetAPIName:     "GetHost",
		utils.GetReqIdentify: "host_id",
	},
	utils.ResourceMappingGroup: {
		utils.RecordKey:      "mapping_group",
		utils.GetAPIName:     "GetMappingGroup",
		utils.GetReqIdentify: "mapping_group_id",
		utils.ListAPIName:    "ListMappingGroups",
		utils.CreateAPIName:  "CreateMappingGroup",
	},
	utils.ResourceNFSGateway: {
		utils.RecordKey:      "nfs_gateway",
		utils.RecordsKey:     "nfs_gateways",
		utils.GetAPIName:     "GetNFSGateway",
		utils.GetReqIdentify: "gateway_id",
		utils.ListAPIName:    "ListNFSGateways",
		utils.CreateAPIName:  "CreateNFSGateway",
	},
	utils.ResourceObjectStorage: {
		utils.RecordKey:     "object_storage",
		utils.GetAPIName:    "GetObjectStorage",
		utils.CreateAPIName: "InitObjectStorage",
	},
	utils.ResourceObjectStorageArchivePool: {
		utils.ListAPIName:    "ListArchivePools",
		utils.GetReqIdentify: "archive_pool_id",
		utils.GetAPIName:     "GetArchivePool",
		utils.RecordKey:      "os_archive_pool",
		utils.RecordsKey:     "os_archive_pools",
		utils.CreateAPIName:  "CreateArchivePool",
	},
	utils.ResourceObjectStorageBucket: {
		utils.ListAPIName:    "ListBuckets",
		utils.GetReqIdentify: "bucket_id",
		utils.GetAPIName:     "GetBucket",
		utils.RecordKey:      "os_bucket",
		utils.RecordsKey:     "os_buckets",
		utils.CreateAPIName:  "CreateBucket",
	},
	utils.ResourceObjectStorageGateway: {
		utils.ListAPIName:    "ListGateways",
		utils.GetAPIName:     "GetGateway",
		utils.GetReqIdentify: "gateway_id",
		utils.RecordKey:      "os_gateway",
		utils.RecordsKey:     "os_gateways",
		utils.CreateAPIName:  "CreateGateway",
	},
	utils.ResourceObjectStoragePolicy: {
		utils.GetReqIdentify: "policy_id",
		utils.GetAPIName:     "GetPolicy",
		utils.ListAPIName:    "ListPolicies",
		utils.RecordKey:      "os_policy",
		utils.RecordsKey:     "os_policies",
		utils.CreateAPIName:  "CreatePolicy",
	},
	utils.ResourceObjectStorageUser: {
		utils.GetReqIdentify: "user_id",
		utils.GetAPIName:     "GetObjectStorageUser",
		utils.ListAPIName:    "ListObjectStorageUsers",
		utils.RecordKey:      "os_user",
		utils.RecordsKey:     "os_users",
		utils.CreateAPIName:  "CreateObjectStorageUser",
	},
	utils.ResourceOsd: {
		utils.GetAPIName:     "GetOsd",
		utils.GetReqIdentify: "osd_id",
		utils.RecordKey:      "osd",
		utils.ListAPIName:    "ListOsds",
		utils.CreateAPIName:  "CreateOsd",
	},
	utils.ResourceOsds: {
		utils.GetAPIName:     "GetOsd",
		utils.GetReqIdentify: "osd_id",
		utils.RecordKey:      "osd",
		utils.ListAPIName:    "ListOsds",
		utils.CreateAPIName:  "CreateOsd",
	},
	utils.ResourcePartitions: {
		utils.CreateAPIName:  "CreatePartitions",
		utils.GetAPIName:     "GetDisk",
		utils.GetReqIdentify: "disk_id",
	},
	utils.ResourcePool: {
		utils.ListAPIName:    "ListPools",
		utils.GetReqIdentify: "pool_id",
		utils.GetAPIName:     "GetPool",
		utils.RecordKey:      "pool",
		utils.RecordsKey:     "pools",
		utils.CreateAPIName:  "CreatePool",
	},
	utils.ResourceS3LoadBalancerGroup: {
		utils.ListAPIName:    "ListS3LoadBalancerGroups",
		utils.GetReqIdentify: "group_id",
		utils.GetAPIName:     "GetS3LoadBalancerGroup",
		utils.RecordKey:      "s3_load_balancer_group",
		utils.RecordsKey:     "s3_load_balancer_groups",
		utils.CreateAPIName:  "CreateS3LoadBalancerGroup",
	},
	utils.ResourceToken: {
		utils.CreateAPIName: "CreateToken",
	},
	utils.ResourceUser: {
		utils.GetAPIName:     "GetUser",
		utils.GetReqIdentify: "user_id",
		utils.RecordKey:      "user",
		utils.RecordsKey:     "users",
		utils.ListAPIName:    "ListUsers",
		utils.CreateAPIName:  "CreateUser",
	},
	utils.ResourceFSUser: {
		utils.GetAPIName:     "GetFSUser",
		utils.GetReqIdentify: "fs_user_id",
		utils.RecordKey:      "fs_user",
		utils.RecordsKey:     "fs_users",
		utils.ListAPIName:    "ListFSUsers",
		utils.CreateAPIName:  "CreateFSUser",
	},
	utils.ResourceFSUserGroup: {
		utils.GetAPIName:     "GetFSUserGroup",
		utils.GetReqIdentify: "fs_user_group_id",
		utils.RecordKey:      "fs_user_group",
		utils.RecordsKey:     "fs_user_groups",
		utils.ListAPIName:    "ListFSUserGroups",
		utils.CreateAPIName:  "CreateFSUserGroup",
	},
	utils.ResourceFSFolder: {
		utils.GetAPIName:     "GetFolder",
		utils.GetReqIdentify: "fs_folder_id",
		utils.RecordKey:      "fs_folder",
		utils.RecordsKey:     "fs_folders",
		utils.ListAPIName:    "ListFolders",
		utils.CreateAPIName:  "CreateFolder",
	},
	utils.ResourceFSClient: {
		utils.GetAPIName:     "GetFSClient",
		utils.GetReqIdentify: "fs_client_id",
		utils.RecordKey:      "fs_client",
		utils.RecordsKey:     "fs_clients",
		utils.ListAPIName:    "ListFSClients",
		utils.CreateAPIName:  "CreateFSClient",
	},
	utils.ResourceFSClientGroup: {
		utils.GetAPIName:     "GetFSClientGroup",
		utils.GetReqIdentify: "fs_client_group_id",
		utils.RecordKey:      "fs_client_group",
		utils.RecordsKey:     "fs_client_groups",
		utils.ListAPIName:    "ListFSClientGroups",
		utils.CreateAPIName:  "CreateFSClientGroup",
	},
	utils.ResourceFSGatewayGroup: {
		utils.GetAPIName:     "GetFSGatewayGroup",
		utils.GetReqIdentify: "fs_gateway_group_id",
		utils.RecordKey:      "fs_gateway_group",
		utils.RecordsKey:     "fs_gateway_groups",
		utils.ListAPIName:    "ListFSGatewayGroups",
		utils.CreateAPIName:  "CreateFSGatewayGroup",
	},
	utils.ResourceFSLdap: {
		utils.GetAPIName:     "GetFSLdap",
		utils.GetReqIdentify: "fs_ldap_id",
		utils.RecordKey:      "fs_ldap",
		utils.RecordsKey:     "fs_ldaps",
		utils.ListAPIName:    "ListFSLdaps",
		utils.CreateAPIName:  "CreateFSLdap",
		utils.StatusKey:      "ActionStatus",
	},
	utils.ResourceFSAD: {
		utils.GetAPIName:     "GetFSActiveDirectory",
		utils.GetReqIdentify: "fs_active_directory_id",
		utils.RecordKey:      "fs_active_directory",
		utils.RecordsKey:     "fs_active_directories",
		utils.ListAPIName:    "ListFSActiveDirectories",
		utils.CreateAPIName:  "CreateFSActiveDirectory",
		utils.StatusKey:      "ActionStatus",
	},
	utils.ResourceFSNFSShare: {
		utils.GetAPIName:     "GetFSNFSShare",
		utils.GetReqIdentify: "fs_nfs_share_id",
		utils.RecordKey:      "fs_nfs_share",
		utils.RecordsKey:     "fs_nfs_shares",
		utils.ListAPIName:    "ListFSNFSShares",
		utils.CreateAPIName:  "CreateFSNFSShare",
	},
	utils.ResourceFSFTPShare: {
		utils.GetAPIName:     "GetFSFTPShare",
		utils.GetReqIdentify: "fs_ftp_share_id",
		utils.RecordKey:      "fs_ftp_share",
		utils.RecordsKey:     "fs_ftp_shares",
		utils.ListAPIName:    "ListFSFTPShares",
		utils.CreateAPIName:  "CreateFSFTPShare",
	},
	utils.ResourceFSSMBShare: {
		utils.GetAPIName:     "GetFSSMBShare",
		utils.GetReqIdentify: "fs_smb_share_id",
		utils.RecordKey:      "fs_smb_share",
		utils.RecordsKey:     "fs_smb_shares",
		utils.ListAPIName:    "ListFSSMBShares",
		utils.CreateAPIName:  "CreateFSSMBShare",
	},
	utils.ResourceNetworkAddress: {
		utils.RecordsKey:  "network_addresses",
		utils.ListAPIName: "ListNetworkAddresses",
	},
	utils.ResourceFSQuotaTree: {
		utils.RecordsKey:     "fs_quota_trees",
		utils.RecordKey:      "fs_quota_tree",
		utils.GetAPIName:     "GetQuotaTree",
		utils.GetReqIdentify: "fs_quota_tree_id",
		utils.ListAPIName:    "ListQuotaTrees",
		utils.CreateAPIName:  "AddFSQuotaTrees",
	},
	utils.ResourceFSArbitrationPool: {
		utils.RecordKey:     "fs_arbitration_pool",
		utils.CreateAPIName: "CreateFSArbitrationPool",
	},
}
