package utils

// list of resource types
const (
	ResourceUser             = "User"
	ResourceToken            = "Token"
	ResourceBootNode         = "BootNode"
	ResourceService          = "Service"
	ResourceHost             = "Host"
	ResourceOsd              = "Osd"
	ResourcePool             = "Pool"
	ResourceBlockVolume      = "BlockVolume"
	ResourceProtectionDomain = "ProtectionDomain"
	ResourceNetworkAddress   = "NetworkAddress"

	ResourceObjectStorage            = "ObjectStorage"
	ResourceObjectStorageUser        = "ObjectStorageUser"
	ResourceObjectStoragePolicy      = "ObjectStoragePolicy"
	ResourceObjectStorageBucket      = "ObjectStorageBucket"
	ResourceObjectStorageGateway     = "ObjectStorageGateway"
	ResourceNFSGateway               = "NFSGateway"
	ResourceObjectStorageArchivePool = "ObjectStorageArchivePool"
	ResourceS3LoadBalancer           = "S3LoadBalancer"
	ResourceS3LoadBalancerGroup      = "S3LoadBalancerGroup"

	ResourceClientGroup  = "ClientGroup"
	ResourceAccessPath   = "AccessPath"
	ResourceMappingGroup = "MappingGroup"

	ResourceFSUser            = "FSUser"
	ResourceFSUserGroup       = "FSUserGroup"
	ResourceFSClient          = "FSClient"
	ResourceFSClientGroup     = "FSClientGroup"
	ResourceFSFolder          = "FSFolder"
	ResourceFSGatewayGroup    = "FSGatewayGroup"
	ResourceFSLdap            = "FSLdap"
	ResourceFSAD              = "FSActiveDirectory"
	ResourceFSNFSShare        = "FSNfsShare"
	ResourceFSFTPShare        = "FSFtpShare"
	ResourceFSSMBShare        = "FSSmbShare"
	ResourceFSQuotaTree       = "FSQuotaTree"
	ResourceFSArbitrationPool = "FSArbitrationPool"

	// a collection of resources of same kind
	ResourceHosts        = "Hosts"
	ResourceOsds         = "Osds"
	ResourceBlockVolumes = "BlockVolumes"
	ResourcePartitions   = "Partitions"

	// logic list of resource used for query or calculation
	ResourceHostList    = "HostList"
	ResourceDiskList    = "DiskList"
	ResourceIntegerList = "IntegerList"
	ResourceStringList  = "StringList"
	ResourceTemplate    = "Template"
)

// osd roles
const (
	PoolOsdRoleData  = "data"
	PoolOsdRoleIndex = "index"
)

// define required parameters
const (
	ParamClusterURL = "ClusterURL"
)

// XmsHeaderAuthToken defines header of xms auth token
const XmsHeaderAuthToken = "Xms-Auth-Token"

// Resource Status
const (
	StatusActive         = "active"
	StatusError          = "error"
	StatusVerifyingError = "verifying_error"
	StatusSyncingError   = "syncing_error"
	StatusUninitialized  = "uninitialized"
	StatusFinished       = "finished"
	StatusHealthy        = "healthy"
)

// DefaultCheckInterval defines interval of checking resource status
const DefaultCheckInterval = 3

// DefaultCheckCount defines count of checking resource status
const DefaultCheckCount = 30

// Defines resource setting keys
const (
	GetReqIdentify = "GetReqIdentify"
	ListAPIName    = "ListAPIName"
	GetAPIName     = "GetAPIName"
	CreateAPIName  = "CreateAPIName"
	UpdateAPIName  = "UpdateAPIName"
	RecordKey      = "RecordKey"
	RecordsKey     = "RecordsKey"
	StatusKey      = "StatusKey"
	IdentifyKey    = "IdentifyKey"
)

// Defines consts for tempaltes
const (
	ContextValueActionRange = "range"
	ContextTypeIntegerList  = "IntegerList"
	ContextTypeStringList   = "StringList"
	ContextTypeString       = "String"
	ContextTypeInteger      = "Integer"
	ContextTypeBool         = "Bool"
)
