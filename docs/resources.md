# formation资源说明

- [formation资源说明](#formation资源说明)
  - [IntegerList & StringList](#integerlist--stringlist)
  - [DiskList](#disklist)
  - [Host](#host)
  - [Hosts](#hosts)
  - [BootNode](#bootnode)
  - [Osd](#osd)
  - [Osds](#osds)
  - [Partitions](#partitions)
  - [Pool](#pool)
  - [ObjectStorage](#objectstorage)
  - [ObjectStorageUser](#objectstorageuser)
  - [ObjectStoragePolicy](#objectstoragepolicy)
  - [ObjectStorageArchivePool (4.1及以上版本不支持)](#objectstoragearchivepool-41及以上版本不支持)
  - [ObjectStorageBucket](#objectstoragebucket)
  - [ObjectStorageGateway](#objectstoragegateway)
  - [NFSGateway](#nfsgateway)
  - [S3LoadBalancerGroup](#s3loadbalancergroup)
  - [MappingGroup](#mappinggroup)
  - [ClientGroup](#clientgroup)
  - [BlockVolume](#blockvolume)
  - [BlockVolumes](#blockvolumes)
  - [AccessPath](#accesspath)
  - [NetworkAddress](#networkaddress)
  - [FSClient](#fsclient)
  - [FSClientGroup](#fsclientgroup)
  - [FSUser](#fsuser)
  - [FSUserGroup](#fsusergroup)
  - [FSArbitrationPool](#fsarbitrationpool)
  - [FSGatewayGroup](#fsgatewaygroup)
  - [FSFolder](#fsfolder)
  - [FSQuotaTree](#fsquotatree)
  - [FSFtpShare](#fsftpshare)
  - [FSNfsShare](#fsnfsshare)
  - [FSSmbShare](#fssmbshare)
  - [FSActiveDirectory](#fsactivedirectory)
  - [FSLdap](#fsldap)

## IntegerList & StringList

```
{
    "Description" : "this template will generate an integer list.",
    "Parameters" : {
        "ClusterURL": {
            "Type" : "String",
            "Value" : "http://10.0.0.1:8056/v1"
        },
        "InputIntegerList": {
            "Type" : "IntegerList",
            "Value" : [1,2]
        }
    },
    "Resources" : [
        {
            "Name" : "array",
            "Type" : "IntegerList",
            "Properties" : {
                "Attributes" : [
                    {"Ref": "InputIntegerList"},
                    [3,4],
                    5
                ]
            }
        }
    ]
}
```

## DiskList

|字段|类型|描述|可关联类型|
|-|-|-|-|
|Used| Bool |是否被占用，创建osd| Bool|
|Device |String |盘符（例如：sda） |String|
|DiskType| String| 盘类型（例如：SSD、HDD）| String|
|Model| String| 盘厂商，支持包含匹配（例如：TOSHIBA、INTEL）| String|
|MinSizeGB| Integer| 盘最小容量，单位GB| Integer|
|MaxSizeGB |Integer |盘最大容量，单位GB |Integer|
|HostIDs| IntegerList |所属服务器列表| Integer,Hosts|
|Num |Integer| 获取Disk的最大数目| Integer|
|NumPerHost |Integer |每个节点上获取Disk的最大数目| Integer|
|IsCache| Bool| 是否为缓存盘| Bool|
|WWID |String |盘WWID，支持包含匹配 |String|
Update 操作支持对盘的某些属性做批量更新，支持的属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|DiskType |String| 盘类型 |String|
|LightingStatus |String| 盘点灯状态| String|

## Host

只支持 Create 操作，默认操作为 Create。在集群中添加新的服务器，Create操作支持的属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|AdminIP| String| 服务器管理ip |String|
|Roles |StringList| 服务器角色列表（例如：admin、monitor、block_storage_gateway、nfs_gateway、s3_gateway、file_storage_gateway) |StringList|
|Type| String |服务器类型("storage_server", "storage_client", "storage_witness")| String|

## Hosts

用来创建或获取host列表,支持 Create,Get 操作，默认操作为 Create。在集群中批量添加新的服务器，Create操作支持的属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|AdminIP |String |服务器管理ip列表| String|
|Roles |StringList| 服务器角色列表（例如：admin、monitor、|block_storage_gateway、nfs_gateway、s3_gateway、file_storage_gateway) |StringList|
|Type |String| 服务器类型("storage_server", "storage_client", "storage_witness")| String|
Get操作用来获取服务器列表，支持指定以下参数选择特定服务器
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Roles| StringList| 服务器角色列表（例如：admin、monitor、block_storage_gateway、nfs_gateway、s3_gateway、file_storage_gateway) |StringList|
|Type |String |服务器类型("storage_server", "storage_client", "storage_witness")| String|

## BootNode

支持 Create 操作，默认操作为 Create。Create操作用于设置集群的网络配置等信息，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|PublicNetwork |String| public网络默认网段（例如：10.0.0.1/24） |String|
|PrivateNetwork| String| private网络默认网段（例如：10.0.1.1/24）|String|
|InstallerPath| String |安装包位置，可选（例如：/opt/sds/installer/sds-release）| String|
|GatewayNetwork |String| gateway网络默认网段（例如：10.0.2.1/24）| String|
|AdminNetwork| String |admin网络默认网段（例如：10.0.3.1/24）| String|

## Osd

支持 Create 操作，默认操作为 Create。Create操作用于创建新的 osd，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|DiskID| Integer| 数据盘 ID |Integer|
|PartitionID |Integer| 缓存盘分区 ID，可选 |Integer|
|Role| String| osd 角色（可选：index、data、compound(仅EOS可用)），可选 |String|
|OmapByte| Integer| omap分区大小(创建对象复合盘时使用)| Integer|

如果设置了PartitionID，表示创建缓存盘。需要注意的是，缓存盘和数据盘需要在同一个服务器上。

## Osds

支持 Create 操作，默认操作为 Create。Create操作用于批量创建新的 osd数组，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|DiskIDs| IntegerList |盘 ID 数组| IntegerList, DiskList|
|PartitionIDs |IntegerList| 缓存盘分区 ID 数组，可选| IntegerList,Partitions|
|Role| String |osd 角色（例如：index、data、compound(仅EOS可用)），可选 |String|
|OmapByte| Integer| omap分区大小(创建对象复合盘时使用)| Integer|

需要注意的是，如果创建缓存盘，DiskIDs 和 PartitionIDs 数组的长度应该一样，在同一个服务器上。

## Partitions

支持 Create 操作，默认操作为 Create。Create操作用于创建新的磁盘分区列表，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|DiskIDs| IntegerList |盘 ID 数组| DiskList,IntegerList|
|NumPerDisk| Integer |单块盘上创建的分区数| Integer|

## Pool

支持 Create 操作，默认操作为 Create。Create操作用于创建新的 pool，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name| String| 池名称| String|
|PoolName |String| ceph中存储池名称| String|
|OsdIDs| IntegerList |osd ID 数组 |IntegerList,Osds|
|PoolType| String| 池类型（例如：replicated、erasure）| String|
|PoolRole| String| 池角色（例如：data、index、compound(仅EOS可用)）| String|
|Size| Integer |副本池副本数，可选| Integer|
|CodingChunkNum |Integer |EC池校验块数，可选| Integer|
|DataChunkNum |Integer |EC池数据块数，可选| Integer|
|FailureDomainType| String| 故障域级别，可选，默认host| String|

## ObjectStorage

支持 Create 操作，默认操作为 Create。Create操作用于初始化对象存储环境，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 池名称 |String|
|PoolID |Integer |索引池id| Integer,Pool|
|ArchivePoolID |Integer| 归并池ID(4.1及以后版本不需配置)| Integer,Pool|

## ObjectStorageUser

支持 Create 操作，默认操作为 Create。Create操作用于新的对象存储用户，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 用户名称| String|
|Email |String| 用户email| String|
|MaxBuckets| Integer| 存储桶配额| Integer|
|BucketQuotaMaxObjects| Integer |单存储桶对象数| Integer|
|BucketQuotaMaxSize| Integer |但存储桶容量 |Integer|
|UserQuotaMaxObjects| Integer |对象数配额| Integer|
|UserQuotaMaxSize |Integer| 容量配额| Integer|
|Keys| ObjectList| 对象用户秘钥列表(字段说明见下表)| -|

|字段|类型|描述|可关联类型|
|-|-|-|-|
|AccessKey |String| ak |String|
|SecretKey |String| sk| String|

## ObjectStoragePolicy

支持 Create 操作，默认操作为 Create。Create操作用于创建新的对象存储策略，属性如下:
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 策略名称 |String|
|Description |String| 策略描述| String|
|IndexPoolId |Integer| 索引池ID| Integer,Pool|
|CachePoolId |Integer| 缓存池ID(4.1以后版本支持)| Integer,Pool|
|DataPoolID |Integer |活动存储池id(4.1以后版本建议使用StorageClass)| Integer,Pool|
|DataPoolIDs |Integer| 存储池id列表(4.1以后版本建议使用StorageClass)| IntegerList|
|Shared Bool| 是否可共享(EOS可用)| Bool|
|Compress| Bool |是否开启压缩| Bool|
|Crypto| Bool| 是否开启加密| Bool|
|StorageClasses| ObjectList| 存储类别配置(目前只支持一个存储类别)，字段说明见下表| -|

|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String |存储类别名| String|
|Description |String| 存储类别ID| String|
|ActivePoolIDs |IntegerList |活动池ID列表| IntegerList|
|InactivePoolIDs| IntegerList |非活动池ID列表 |IntegerList|

## ObjectStorageArchivePool (4.1及以上版本不支持)

|字段|类型|描述|可关联类型|
|-|-|-|-|
|PoolID| Integer |存储池id |Integer|

## ObjectStorageBucket

|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 桶名称| String|
|AllUserPermission| String| 匿名用户权限(read:读权限、write:写权限、*：所有权限，"": 无权限，多种权限使用,连接) |String|
|AuthUserPermission| String| 认证用户权限(read:读权限、write:写权限、*：所有权限，"": 无权限，多种权限使用,连接) |String|
|OwnerID| Integer| 桶所有者ID| Integer,ObjectStorageUser|
|OwnerPermission |String| 桶所有者权限(read:读权限、write:写权限、*：所有权限，"": 无权限，多种权限使用,连接) | String|
|PolicyID| Integer| 存储策略ID| Integer,ObjectStoragePolicy|
|QuotaMaxObjects |Integer |对象数配额| Integer|
|QuotaMaxSize |Integer| 容量配额| Integer|

## ObjectStorageGateway

支持 Create 操作，默认操作为 Create。Create操作用于创建新的对象存储网关，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 网关名称| String|
|Description |String| 网关描述| String|
|HostID| Integer| 网关对应Host| id(s3_gateway角色) |Integer,Host|
|GatewayIP| String| 网关ip |String|
|Port |Integer| 网关监听端口| Integer|

## NFSGateway

支持 Create 操作，默认操作为 Create。Create操作用于创建新的对象存储NFS网关，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String| 网关名称| String|
|Port| Integer| 网关描述 |String|
|HostID| Integer |网关对应Host |id(s3_gateway角色) |Integer,Host|
|GatewayIP| String| 网关ip| String|
|Port| Integer| 网关监听端口 |Integer|

## S3LoadBalancerGroup

支持 Create 操作，默认操作为 Create。Create操作用于创建新的对象路由，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name |String |对象路由名称| String|
|Port |Integer |对象路由服务监听端口 |Integer|
|S3LoadBalancers| ObjectList| 负载均衡器配置，字段如下表| -|

|字段|类型|描述|可关联类型|
|-|-|-|-|
|HostID| Integer |负载均衡器对应Host| Integer,Host|
|VIP| String| 负载均衡器vip| String|

## MappingGroup

支持 Create 操作，默认操作为 Create。Create操作用于创建新的映射组，属性如下：  
|字段|类型|描述|可关联类型|
|-|-|-|-|
|AccessPathID| Integer| 访问路径ID| Integer|
|BlockVolumeIDs| IntegerList| 卷ID列表| IntegerList,BlockVolumes|
|ClientGroupID |Integer| 客户端组ID| Integer，ClientGroup|

## ClientGroup

 支持 Create 操作，默认操作为 Create。Create操作用于创建新的客户端组，属性如下：  
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name| String| 名称| String|
|Type| String| 类型(可选："iSCSI","Kubernetes","FC")| String|
|Description| String| 描述| String|
|Clients |ObjectList| 客户端列表，字段如下表| -|

|字段|类型|描述|可关联类型|
|-|-|-|-|
|Code| String |客户端iqn/ip/wwn |String|

## BlockVolume

支持 Create 操作，默认操作为 Create。Create操作用于创建新的块存储卷，属性如下：  
|字段|类型|描述|可关联类型|
|-|-|-|-|
|Name| String| 名称 |String|
|Description| String| 描述 |String
|BlockSnapshotID |Integer| 快照id| Integer|
|PerformancePriority| Integer |性能优先级(0: 默认 1:优先)| Integer|
|PoolID| Integer |存储池id |Integer,Pool|
|Size| Integer| 卷大小| Integer|
|QosEnabled |Bool| 是否开启qos| Bool|
|Qos |Object |qos配置信息，字段说明如下表| -|

|字段|类型|描述|可关联类型|
|-|-|-|-|
|BurstTotalBw |Integer |突发带宽| Integer|
|BurstTotalIops| Integer |突发IOPS |Integer|
|MaxTotalBw |Integer|最大带宽 |Integer|
|MaxTotalIops |Integer| 最大IOPS| Integer|

## BlockVolumes

支持 Create 操作，默认操作为 Create。Create操作用于批量创建新的块存储卷，属性如下：  
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Names | StringList | 名称列表，由提供的卷名数量决定创建数量 | StringList |
| Prefix | String | 批量创建卷名前缀 | String |
| Num | Integer | 批量创建数量(配合前缀使用) | Integer |
| Description | String | 描述 | String |
| BlockSnapshotID | Integer | 快照id | Integer |
| PerformancePriority | Integer | 性能优先级(0: 默认 1:优先) | Integer |
| PoolID | Integer | 存储池id | Integer,Pool |
| Size | Integer | 卷大小 | Integer |
| QosEnabled | Bool | 是否开启qos | Bool |
| Qos | Object | qos配置信息，字段说明如下表 | - |

|字段|类型|描述|可关联类型|
|-|-|-|-|
| BurstTotalBw | Integer | 突发带宽 | Integer |
| BurstTotalIops | Integer | 突发IOPS | Integer |
| MaxTotalBw | Integer | 最大带宽 | Integer |
| MaxTotalIops | Integer | 最大IOPS | Integer |

## AccessPath

支持 Create 操作，默认操作为 Create。Create操作用于创建新的访问路径，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 名称列表，由提供的卷名数量决定创建数量 | String |
| Description | String | 批量创建卷名前缀 | String |
| Chap | Bool | Chap | Bool |
| Tname | String | 目标端名称 | String |
| Tsecret | String | 目标端机密 | String |
| Type | String | 类型(可选："iSCSI", "FC", "Local") | String |
| MappingGroups | ObjectList | 映射组配置列表，字段内容参看 MappingGroup | - |

## NetworkAddress

支持 Get 操作，无默认操作。Get操作用于获取指定ip的id，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
|IP| String |ip地址| String|

## FSClient

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件客户端，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name |String |客户端名| String |
| IP |String|客户端网段或IP。网段如：10.0.0.1/24|String |

## FSClientGroup

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件客户端组，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 客户端组名 | String |
| ClientIDs | IntegerList | 客户端id列表 | IntegerList |

## FSUser

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件用户，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 文件用户名 | String |
| Email | String | 用户邮箱 | String |
| Password | String | 用户密码 | String |

## FSUserGroup

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件用户组，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 文件用户组名 | String |
| UserIDs | IntegerList | 文件用户id列表 | IntegerList |

## FSArbitrationPool

支持 Create 操作，默认操作为 Create。Create操作用于创建的文件网关高可用仲裁池，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| PoolID | Integer | 仲裁池对应存储池id | Integer,Pool |

## FSGatewayGroup

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件网关组，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 文件网管组名 | String |
| Description | String | 文件网管组描述 | String |
| VIP | String | 网管组vip | String |
| Types | StringList | 网管组类型(可选：smb | nfs | ftp) | StringList |
| Security | String | 网管组认证类型(可选：local | ad | ldap) | String |
| SMB1Enabled | Bool | 是否启用SMB1.0 | Bool |
| SMBPorts | IntegerList | SMB协议端口列表 | IntegerList |
| NFSVersions | StringList | Nfs版本列表(可选"v3", "v4", "v4.1", "v4.2") | StringList |
| Encoding | String | 编码方式(可选utf8 gbk) | String |
| Gateways | ObjectList | 网关列表 | 字段说明如下表 |

|字段|类型|描述|可关联类型|
|-|-|-|-|
| HostID | Integer | 网关节点id | Integer,Host |
| NetworkAddressID | Integer | 网关节点地址id | Integer,NetworkAddress |

## FSFolder

支持 Create 操作，默认操作为 Create。Create操作用于创建新的文件系统，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 文件系统名 | String |
| Description | String | 文件系统描述 | String |
| PoolID | Integer | 存储池id | Integer,Pool |
| Size | Integer | 文件系统容量(字节) | Integer |

## FSQuotaTree

支持 Create 操作，默认操作为 Create。Create操作用于创建新的QuotaTree，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | QuotaTree名 | String |
| FolderID | Integer | 文件系统id | Integer, FSFolder |
| SoftQuotaSize | Integer | 软配额 | Integer |
| Size | Integer | 硬配额 | StringList |

## FSFtpShare

支持 Create 操作，默认操作为 Create。Create操作用于创建新的ftp共享，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 共享名 | String |
| FolderID | Integer | 文件系统id | Integer,FSFolder |
| QuotaTreeID | Integer | QuotaTree id | Integer,FSQuotaTree |
| GatewayGroupID | Integer | 网管组id | Integer,FSGatewayGroup |
| ACLs | ObjectList | 访问名单配置 | 字段说明如下表 |

|字段|类型|描述|可关联类型|
|-|-|-|-|
| Type | String | 用户组类型(可选:user_group,anonymous) | String |
| UserGroupID | Integer | 用户组id(Type为user_group时指定) | Integer,FSUserGroup |
| ListEnabled | Bool | 是否可以查看文件列表 | Bool |
| CreateEnabled | Bool | 是否可以创建文件夹 | Bool |
| RenameEnabled | Bool | 是否可以重命名 | Bool |
| DeleteEnabled | Bool | 是否可以删除 | Bool |
| UploadEnabled | Bool | 是否可以上床 | Bool |
| DownloadEnabled | Bool | 是否可以下载 | Bool |

## FSNfsShare

支持 Create 操作，默认操作为 Create。Create操作用于创建新的nfs共享，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| FolderID | Integer | 文件系统id | Integer,FSFolder |
| QuotaTreeID | Integer | QuotaTree id | Integer,FSQuotaTree |
| GatewayGroupID | Integer | 网管组id | Integer, FSGatewayGroup |
| ACLs | ObjectList | 访问名单配置 | 字段说明如下表 |

|字段|类型|描述|可关联类型|
|-|-|-|-|
| Type | String | 用户组类型(可选:client_group,every_client) | String |
| ClientGroupID | Integer | 客户端组id(Type为client_group时指定) | Integer,FSClientGroup |
| Permission | String | nfs权限(RO  RW) | String |
| Sync | Bool | 是否同步写入模式 | Bool |
| AllSquash | Bool | 是否开启AllSquash | Bool |
| RootSquash | Bool | 是否开启RootSquash | Bool |

## FSSmbShare

支持 Create 操作，默认操作为 Create。Create操作用于创建新的smb共享，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | smb共享名 | String |
| FolderID | Integer | 文件系统id | Integer,FSFolder |
| QuotaTreeID | Integer | QuotaTree id | Integer,FSQuotaTree |
| GatewayGroupID | Integer | 网管组id | Integer, FSGatewayGroup |
| Recycled | Bool | 启用回收站 | Bool |
| ACLInherited | Bool | 添加默认ACL | Bool |
| CaseSensitive | Bool | 大小写敏感 | Bool |
| ACLs | ObjectList | 访问名单配置 | 字段说明如下表 |

|字段|类型|描述|可关联类型|
|-|-|-|-|
| Type | String | 用户组类型(可选:user_group,everyone) | String |
| UserGroupID | Integer | 客户端组id(Type为user_group时指定) | Integer,FSUserGroup |
| UserGroupName | String | 用户组名 | String |
| Permission | String | 权限(RO  RW  full_control) | String |

## FSActiveDirectory

支持 Create 操作，默认操作为 Create。Create操作用于添加AD域，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 名称 | String |
| Workgroup | Stinrg | 域名 | Stinrg |
| Realm | String | 域 | Stinrg |
| IP | String | DNS | IP | Stinrg |
| UserName | Stinrg | 用户名 | Stinrg |
| Password | Stinrg | 密码 | Stinrg |

## FSLdap

支持 Create 操作，默认操作为 Create。Create操作用于添加ldap域，属性如下：
|字段|类型|描述|可关联类型|
|-|-|-|-|
| Name | String | 名称 | String |
| IP | Stinrg | 服务器IP | Stinrg |
| Port | Integer | 端口号 | Integer |
| Suffix | Stinrg | 基准DN | Stinrg |
| AdminDN | Stinrg | 绑定DN | Stinrg |
| Password | Stinrg | 绑定密码 | Stinrg |
| UserSuffix | Stinrg | 用户所在目录 | Stinrg |
| GroupSuffix | Stinrg | 用户组所在目录 | Stinrg |
| Timeout | Integer | 查询超时时间 | Integer |
| ConnectionTimeout | Integer | 连接超时时间 | Integer |
