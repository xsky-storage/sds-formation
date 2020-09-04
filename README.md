# sds-formation

## 编译

使用如下命令会在项目目录的"bundles/latest/binary/sds-formation"位置构建sds-formation的可执行文件

```
make binary
```

## 运行单元测试

```
make unit-test
```

## 使用说明

sds-formation 提供了一种通用语言来描述和预配置 SDS产品的所有存储资源。用户可以通过编辑 json 格式的预配置文件，设置集群的配置参数，从而安装指定的模板，有序创建和更新存储集群中的相关资源。  
json 预配置文件包括三个部分：

- Description: 预配置文件说明，可选
- Parameters: 可变参数列表
- Resources: 待配置的资源

一个简单的json配置文件示例如下：

```
{
    "Description" : "this template will create a host with given admin IP.",
    "Parameters" : {
        "ClusterURL" : {
            "Type" : "String",
            "Value" : "http://10.0.0.1:8056/v1"
        },
        "admin_ip" : {
            "Type" : "String",
            "Value" : "172.16.31.119"
        }
    },
    "Resources" : [
        {
            "Name": "t1",
            "Type" : "Token",
            "Action" : "Create",
            "Properties" : {
                "Name" : "admin",
                "Password" : "admin"
            }
        },
        {
            "Name" : "h1",
            "Type" : "Host",
            "Action" : "Create",
            "WaitInterval" : 100,
            "CheckInterval" : 5,
            "Properties" : {
                "AdminIP" : {"Ref": "admin_ip"}
            }
        }
    ]
}
```

其中：

- Description 表述了通过该模板的作用
- Parameters 中包含了两个可变参数：ClusterURL、admin_ip
  - - ClusterURL 是 formation 是系统中的默认参数，用以表示将要操作的存储集群的 API 入口。ClusterURL 类型必须要为 String，模板中的具体值为 "http://10.0.0.1:8056/v1"
  - - admin_ip 是用户自定义的参数，类型为String，值为 “172.16.31.119”。用户自定义参数可以在模板的 Resources 部分中使用。
- Resources 部分创建了两个资源 t1 和 h1。其中，t1 是 Token 类型的临时 token 资源。在存储集群初始化完成之后，会默认开启 token 认证，对存储资源的操作需要持有token。formation 默认是安装顺序执行操作的，所以在执行该脚本时，会首先在目标存储集群中创建临时 token，并在之后，持有该 token 执行后续操作。h1 为 Host 类型创建操作。由于该操作是异步操作，所以还设置了资源的检查等待间隔 CheckInterval 为100秒，100秒之后开始检查资源状态，每次检查的间隔 CheckInterval 为 5 秒。该操作只包含了一个属性 AdminIP。Admin 赋值为 Parameters 中的 admin_ip 变量，注意这里使用到了一个函数操作 Ref。

### 函数

#### Ref 引用函数

初始的示例中在创建 h1 host时，设置 AdminIP 的值为 Parameters 中声明的 AdminIP。这里通过使用 Ref 函数，引用了资源 admin_ip 的值。需要说明的是，Ref 除了可以引用 Parameters 中声明的变量以外，也可以引用模板顺序创建过程中生成（Get 或者 Create）的资源  

#### Select 数组选择函数

select 数组选择函数，用于在 String 数组或者 Integer 数组中选择元素。在使用 Select 函数时，通常会同时使用 Ref 函数，用于选择之前创建或者获取的数组资源。

### 资源

sds-formation 中的 Resource 共支持 6 个字段：

- Name：资源名称，仅限于模板中使用，与存储集群无关，资源名称最好唯一，否则会被覆盖。
- Type: 资源的类型，包括唯一资源<Resource>，和数组资源"<Resource>s"、"<Resource>List"。其中只支持 Get 操作的数组资源以 List 结尾，如 DiskList。
- Action: 操作类型，包括Get、Create、Update，目前还不支持Delete操作。需要注意的是，Action 为可选参数，如果未设置则使用相应资源的默认操作。
- WaitInterval: 资源状态检查开始的等待间隔。对于异步操作，可以通过调整资源检查开始的等待间隔，来适配不同环境的资源创建速度。单位为秒，如果未设置，则不等待立刻开始周期性检查。
- CheckInterval: 资源状态的检查间隔。对于异步资源，formation会定期检查资源的状态是否正常，最大检查次数是30次。单位为秒，如果未设置或者设置为0，则使用相应资源的默认检查间隔，通常为5秒，部分创建时间较长的资源和批量资源做了调整。
- Properties: 资源的属性，具体包括哪些属性与Type和Action的值有关。
具体的支持的资源类型可以参考[资源说明](./docs/resources.md)

### 模板

formation目前已经支持模板功能，模板时一段通用的资源创建配置信息，模板配置信息写在formation json文件Templates字段下,与创建资源的Resources同级，如下是一个典型的模板示例,Templates值中key为模板名，值为模板内容，模板内容的写法与一般资源创建的下发基本相同：

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
            "Value": ["10.0.0.1"]
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

关于模板的具体使用示例可以参考[模板说明](./docs/template.md)

### 其他功能说明

1.dry-run  
运行formation时可以通过指定-dry-run来对json文件做基本检测，此模式下并不会实际去创建资源

2.可重入  
为formation运行过程添加了缓存机制，对于创建成功的资源会记录其标识信息，在某次运行中断时可以在下次运行时继续运行，同时额外说明如下：

- 缓存信息默认保存在当前目录的foramtion_cache文件夹下，可通过-cache-path选项自定义缓存文件目录
- 缓存文件由json中的描述及集群url唯一标识
- 可以通过在运行时指定-no-continue来不适用缓存(同时会删除已存在缓存文件)
- 在运行中断后，可以修改还未创建资源的信息，但是不能修改已创建资源的信息
