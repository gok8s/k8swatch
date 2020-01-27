package event

//可不报警，仅做警示。
var WarnAlertReasonType = map[string]string{
	"Unhealthy":               "容器Health接口异常，状态为Unhealthy",
	"FailedSync":              "FailedSync",
	"FailedPostStartHook":     "容器创建时，执行初始化脚本出错",
	"FailedPreStopHook":       "容器销毁时，执行脚本出错，容器会被正常销毁，需要修改PreStop shell",
	"ImageGCFailed":           "ImageGCFailed",
	"FailedDaemonPod":         "FailedDaemonPod",
	"NodeControllerEviction":  "标记node上的pod为待删除",
	"NodeAllocatableEnforced": "NodeAllocatableEnforced", //调整节点的资源配额
	"ProbeWarning":            "ProbeWarning 健康检查有异常",    //k8s1.14 新增不支持跨主机跳转的healthCheck，会统一认为healthy并发送该警告事件。
	"UnfinishedPreStopHook":   "容器在销毁前执行用户自定义脚本超过用户设置的TimeOut，容器会被强行删除，请配置正确的Shell脚本",
}

var UserAlertReasonType = map[string]string{
	"HostPortConflict": "Host节点端口冲突，请配置正确的端口",
	"BackOff":          "BackOff",
	"CrashLoopBackOff": "容器启动失败",
	"Killing":          "容器被删除",
	"UnhealthKilling":  "容器因健康检查失败被删除", //diy
}
var AdminAlertReasonType = map[string]string{
	"SystemOOM":                    "节点系统发生oom",
	"FailedCreatePodContainer":     "创建容器失败",
	"Evicted":                      "Pod因节点异常被驱逐",
	"DeletingAllPods":              "删除节点上所有pod",
	"FailedCreate":                 "FailedCreate",
	"ReplicaSetCreateError":        "ReplicaSetCreateError",
	"EvictionThresholdMet":         "磁盘空间不足准备尝试清理",
	"FailedToStartNodeHealthcheck": "FailedToStartNodeHealthcheck",
	"NodeHasDiskPressure":          "kubelet面临节点磁盘可用空间不足的压力",
	"NodeNotReady":                 "Node节点不可用",
	"NodeHasInsufficientMemory":    "Node没有足够的可用内存",
	"InvalidDiskCapacity":          "InvalidDiskCapacity ",
	"FreeDiskSpaceFailed":          "FreeDiskSpaceFailed",
	"InsufficientFreeCPU":          "没有足够的CPU",
	"InsufficientFreeMemory":       "没有足够的内存",
	"HostNetworkNotSupported":      "HostNetworkNotSupported",
	"Failed":                       "Failed", //FailedToPullImage       =  需要过滤FailedPullImages，Failed to pull image
	"NodeNotSchedulable":           "Node不可被调度",
	"KubeletSetupFailed":           "KubeletSetupFailed",
	"FailedAttachVolume":           "FailedAttachVolume",
	"FailedDetachVolume":           "FailedDetachVolume",
	"VolumeResizeFailed":           "VolumeResizeFailed",
	"FileSystemResizeFailed":       "FileSystemResizeFailed",
	"FailedUnMount":                "FailedUnMount",
	"FailedUnmapDevice":            "FailedUnmapDevice",
	"HostPortConflict":             "Host节点端口冲突，请配置正确的端口",
	"NodeSelectorMismatching":      "NodeSelectorMismatching",
	"NilShaper":                    "NilShaper",
	"Rebooted":                     "Rebooted",
	"ContainerGCFailed":            "ContainerGCFailed",
	"ErrImageNeverPull":            "ErrImageNeverPull",
	"NetworkNotReady":              "NetworkNotReady",
	"FailedKillPod":                "FailedKillPod",
	"RemovingNode":                 "节点被执行下线",
	"FailedMount":                  "FailedMount",
	"FailedScheduling":             "调度失败",
}
var NormalReasonType = map[string]string{
	"NodeSchedulable":         "NodeSchedulable",
	"NodeReady":               "NodeReady",
	"Pulling":                 "Pulling",
	"Scheduled":               "Scheduled",
	"Pulled":                  "Pulled",
	"Started":                 "Started",
	"Created":                 "Created",
	"CREATE":                  "CREATE",
	"UPDATE":                  "UPDATE",
	"DELETE":                  "DELETE",
	"Starting":                "Starting",
	"SuccessfulMountVolume":   "SuccessfulMountVolume",
	"SuccessfulCreate":        "SuccessfulCreate",
	"SuccessfulDelete":        "SuccessfulDelete",
	"ScalingReplicaSet":       "ScalingReplicaSet",
	"RegisteredNode":          "RegisteredNode",
	"LeaderElection":          "LeaderElection",
	"CreatedLoadBalancer":     "CreatedLoadBalancer",
	"NodeHasNoDiskPressure":   "Node节点没有磁盘压力",
	"NodeHasSufficientMemory": "Node节点有足够的可用内存",
	"NodeHasSufficientDisk":   "Node节点有足够的可用硬盘空间",
	"SandboxChanged":          "SandboxChanged",
	"FailedCreatePodSandBox":  "FailedCreatePodSandBox",
	"FailedPodSandBoxStatus":  "FailedPodSandBoxStatus",
}

var RecoverReasonType = map[string]string{
	"NodeReady": "NodeReady",
}

//TODO 改为可配置的
var AdminAlertNS = map[string]string{
	"kube-system":        "kube-system-ns",
	"kube-public":        "kube-public-ns",
	"default":            "default-ns",
	"ingress-nginx":      "ingress-nginx",
	"ingress-nginx-blue": "ingress-nginx-blue",
	"cre":                "cre",
	"monitoring":         "monitoring",
	"ops":                "ops",
	"weave":              "weave",
}
