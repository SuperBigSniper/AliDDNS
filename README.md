## 0.简要介绍

AliDDNS

## 1.使用说明

> 使用本工具的时候，请详细阅读使用说明。

### 1.1 配置说明

通过更改 ```settings.json.example``` 的内容来实现 DDNS 更新，其文件内部各个选项的说明如下：

```json
{
    "AccessIdComment": "阿里云的 Access Id。",
    "AccessId": "AccessId",
    "AccessKeyComment": "阿里云的 Access Key。",
    "AccessKey": "AccessKey",
    "MainDomainComment": "主域名。",
    "MainDomain": "example.com",
    "SubDomainsComment": "需要批量变更的子域名记录集合。",
    "SubDomains": [
        {
            "TypeComment": "子域名记录类型。",
            "Type": "A",
            "SubDomainComment": "子域名记录前缀。",
            "SubDomain": "sub1",
            "IntervalComment": "TTL 时间。",
            "Interval": 600
        },
        {
            "Type": "A",
            "SubDomain": "sub2",
            "Interval": 600
        }
    ]
}
```

其中 ```Access Id``` 与 ```Access Key``` 可以登录阿里云之后在右上角可以得到。

### 1.2 使用说明

在运行程序的时候，请建立一个新的 ```settings.json``` 文件，在里面填入配置内容，然后执行以下命令：

```shell
./AliDDNS
```

当然如果你有其他的配置文件也可以通过指定 ```-f``` 参数来制定配置文件路径。例如：

```shell
./AliDDNS -f ./settings.json
```
如果你需要启动自动周期检测的话，请通过 `-i` 参数指定执行周期，单位是秒。会自动检测和第一次设置的地址是否相同，相同就不会进行二次设置

```shell
./AliDDNS -f ./settings.json -i 3600
```

> **注意：**
>
> **当你通过 -i 指定了周期之后，请在最末尾使用 & 符号，使应用程序在后台运行。使用群晖的同学，就不要指定 -i 参数了。**

