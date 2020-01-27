## k8swatch 
- k8swatch是监听并存储k8s各类资源变化的一个工具，并能对事件类资源进行分析、报警、查询处理。
- 它通过自定义控制器实现对k8s各种资源的list&watch，并支持以插件化的方式实现对各种资源变更的处理，如存储、发送到MQ、分析报警等。
- 参考开源项目kubewatch并做了改进。

### 功能
- 提供HTTP接口对各类资源的事件进行查询
- alert，用来对events做分析和异常报警，并生成json日志落盘，会被收集到日志平台分析。
- elasticsearch，发布到elasticsearch，支持更久存储和查询
- rabbitmq，发布到rabbitmq,目前graylog会从其消费并做分析报表，如：
    - 应用发布频率统计
    - 业务组发布量排名
    - 应用新建或销毁的记录
    - node的心跳宏观趋势（node的update多为心跳）
- influxdb
- webhook，调用webhook用于后续扩展

#### 已支持的资源类别
- events
- endpoints
- pods
- nodes
- services
- namespaces
- replicationcontrolleres
- persistentvolumes
- persistentvolumeclaim
- secrets
- configmaps
- deployments
- replicasets
- daemonset
- ingresses
- jobs
- roles
- rolebindings
- clusterroles
- clusterrolebindings

#### HTTP查询
- 方法:GET
- 参数:
    - namespace,string,命名空间同k8s
    - kind_name,string,如podname
    - limit,string,最大条数,不设则为10
- 返回:[]{"k":"v"}如
```$xslt
[
{
"count": 340,
"createTimestamp": "2019-10-18 12:01:36",
"evt_name": "asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f",
"firstTimestamp": "2019-10-18 12:01:36",
"kind": "events",
"kind_name": "asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f",
"lastTimestamp": "2019-10-19 23:23:30",
"message": "Readiness probe failed: Get http://10.12.35.133:8080/info: net/http: request canceled (Client.Timeout exceeded while awaiting headers)\nevents:asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f in namespace test-coreasset has been CREATED\n",
"namespace": "test-coreasset",
"reason": "Unhealthy",
"resourceVersion": "",
"source_component": "kubelet",
"source_host": "t-k8s-node-38-99",
"time": "2019-10-19 23:23:30",
"type": "Warning"
},
{
"count": 343,
"createTimestamp": "2019-10-18 12:01:36",
"evt_name": "asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f",
"firstTimestamp": "2019-10-18 12:01:36",
"kind": "events",
"kind_name": "asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f",
"lastTimestamp": "2019-10-19 23:24:46",
"message": "Readiness probe failed: Get http://10.12.35.133:8080/info: net/http: request canceled (Client.Timeout exceeded while awaiting headers)\nevents:asset-restful-8458bb8dcc-bvxjh.15cea1de7fce8b7f in namespace test-coreasset has been UPDATED\n",
"namespace": "test-coreasset",
"reason": "Unhealthy",
"resourceVersion": "",
"source_component": "kubelet",
"source_host": "t-k8s-node-38-99",
"time": "2019-10-19 23:24:46",
"type": "Warning"
}
]
```


### Usage
- ./k8swatch -h查看帮助信息
- 通常./k8swatch --config xxx即可
- 或通过环境变量如：export K8SWATCH_CONF=cretest  ./k8swatch
将读取当前目录下的configs/ 或/etc/k8swatch/configs/或$HOME下对应的yaml结尾的配置文件。

