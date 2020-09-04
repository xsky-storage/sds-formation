# sds-formation模板说明文档

formation目前已经支持模板功能，模板时一段通用的资源创建配置信息，模板配置信息写在formation
json文件Templates字段下,与创建资源的Resources同级，如下是一个典型的模板示例,Templates值中key为模板名，值为模板内容，模板内容
的写法与一般资源创建的下发基本相同：

```
{
    "Description": "example",
    "Parameters": {
        "ClusterURL": {
            "Type": "String",
            "Value": "http://10.0.0.1:8056/v1"
        },
        "Network": {
            "Type": "String",
            "Value": "10.0.0.1/24"
        },
        "AdminIPs": {
            "Type": "StringList",
            "Value": ["10.0.0.2"]
        }
    },
    "Templates": {
        "osdstemplate": [{
            "Name": "PartitionsDisk",
            "Type": "DiskList",
            "Properties": {
                "Used": false,
                "HostIDs": [{
                    "Ref": "host_id"
                }],
                "Num": 1
            }
        }, {
            "Name": "PartitionsDisk",
            "Type": "DiskList",
            "Action": "Update",
            "Properties": {
                "DiskType": "SSD"
            }
        }, {
            "Name": "Partitions",
            "Type": "Partitions",
            "Properties": {
                "DiskIDs": {
                    "Ref": "PartitionsDisk"
                },
                "NumPerDisk": 2
            }
        }, {
            "Name": "diskList",
            "Type": "DiskList",
            "Properties": {
                "Used": false,
                "DiskType": "HDD",
                "HostIDs": [{
                    "Ref": "host_id"
                }],
                "Num": 2
            }
        }, {
            "Name": "osds",
            "Type": "Osds",
            "Properties": {
                "DiskIDs": {
                    "Ref": "diskList"
                },
                "PartitionIDs": {
                    "Ref": "Partitions"
                },
                "Role": "compound"
            }
        }]
    },
    "Resources": [{
        "Name": "Token",
        "Type": "Token",
        "Properties": {
            "Name": "admin",
            "Password": "admin"
        }
    }, {
        "Name": "osds",
        "Type": "Template",
        "TemplateName": "osdstemplate",
        "Context": [{
            "Name": "host_id",
            "Type": "IntegerList",
            "Value": [1, 2],
            "Action": "range"
        }]
    }, {
        "Name": "Pool",
        "Type": "Pool",
        "Properties": {
            "Name": "apool",
            "PoolRole": "compound",
            "Size": 1,
            "PoolType": "replicated",
            "OsdIDs": {
                "TemplateAttr": {
                    "Ref": "osds",
                    "Attr": "osds"
                }
            }
        }
    }]
}
```

在声明模板后，可以在Resources中使用模板创建资源，模板资源的Type为Template,其不需要Properties属性，而需要Context属性来配置调用模板创建资源时的上下文同时需要TemplateName字段配置模板名来说明使用哪个模板创建资源，在运行某个模板时模板内部的Ref及Select函数可以访问到模板上下文中的字段值及全局声明或已创建的资源值。

模板资源的每个上下文有Name(名称，可以用来给Ref/Select函数引用)，Type(类型，目前只支持Integer,IntegerList,String,StringList四种类型)，Value(上下文的值，可以是字面值也可以时Ref引用)，此外上下文还支持Action可选字段，来指定上下文的模式，默认模式下不管上下文值是否为数组类型，都会使用该上下文原始值，action目前除了默认动作外只支持range动作，指定此动作时要求值类型为数组类型，此时会将数组中的每个元素与其他上下文做全组合，每种组合会用来执行一次模板，示例说明如下：

osdtemplate模板如下：

```
[{
 "Name": "PartitionsDisk",
 "Type": "DiskList",
 "Properties": {
  "Used": false,
  "HostIDs": [{
   "Ref": "host_id"
  }],
  "Num": 1
 }
}, {
 "Name": "PartitionsDisk",
 "Type": "DiskList",
 "Action": "Update",
 "Properties": {
  "DiskType": "SSD"
 }
}, {
 "Name": "Partitions",
 "Type": "Partitions",
 "Properties": {
  "DiskIDs": {
   "Ref": "PartitionsDisk"
  },
  "NumPerDisk": 2
 }
}, {
 "Name": "diskList",
 "Type": "DiskList",
 "Properties": {
  "Used": false,
  "DiskType": "HDD",
  "HostIDs": [{
   "Ref": "host_id"
  }],
  "Num": 2
 }
}, {
 "Name": "osds",
 "Type": "Osds",
 "Properties": {
  "DiskIDs": {
   "Ref": "diskList"
  },
  "PartitionIDs": {
   "Ref": "Partitions"
  },
  "Role": "compound"
 }
}]
```

resources中模板资源内容如下：

```
{
        "Name": "osds",
        "Type": "Template",
        "TemplateName": "osdstemplate",
        "Context": [{
            "Name": "host_id",
            "Type": "IntegerList",
            "Value": [1, 2],
            "Action": "range"
        }]
}
```

则osdstemplate会运行两次，每次运行时的host_id分别为1和2，如此在模板资源osds创建完成后，其值就包含在两个host上创建的所有混合盘osd,实际上模板资源在formation内存储方式为一个数组，数组中每个元素都是一个map, 表示模板的一次运行，模板每次运行创建的资源会包含在该map中，此外，如果需要引用模板内创建的资源，则需要使用如下如下两个函数：

TemplateAttr：用于引用模板中某个资源的所有运行结果并以数组的方式返回，如该资源为数组类型，则会将多个数组合并为一个数组返回，使用示例如下：

```
{
    "Name": "Pool",
    "Type": "Pool",
    "Properties": {
        "Name": "pool1",
        "PoolRole": "data",
        "size": 1,
        "pool_type": "replicated",
        "OsdIDs": {"TemplateAttr": {"Ref": "osds", "Attr": "osds"}}
    }
}
```

如上为使用osds模板资源创建的osd创建pool， “Ref”为模板资源名，Attr为要访问的模板中创建的资源名

TemplateAttrElem: 用户引用模板某次运行的结果，以其原始类型返回，使用示例如下：

```
{
    "Name": "Pool",
    "Type": "Pool",
    "Properties": {
        "Name": "pool1",
        "PoolRole": "data",
        "size": 1,
        "pool_type": "replicated",
        "OsdIDs": {"TemplateAttrElem": {"Ref": "osds", "Attr": "osds", Index: 0}}
    }
}
```

如上为使用osds模板资源中第一次运行创建的osd创建pool,各参数含义与TemplateAttr基本相同，额外增加Index参数来说明使用哪次运行的值。

综上，就可以理解一开始给的模板示例为在每个host上创建混合盘，并使用这些混合盘创建一个pool. 
