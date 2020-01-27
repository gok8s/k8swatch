package event

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gok8s/k8swatch/utils"

	"github.com/gok8s/k8swatch/utils/zlog"

	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	rbacV1 "k8s.io/api/rbac/v1"
)

/*
Event represent an event got from k8s api server
Events from different endpoints need to be casted to KubewatchEvent
before being able to be handled by handler

kind 和cache.NewListWatchFromClient使用的resource要一致
*/
type Event struct {
	Namespace               string `json:"namespace"`
	Kind                    string `json:"kind"`
	Component               string `json:"component"`
	Host                    string `json:"host"`
	Reason                  string `json:"reason"`
	Status                  string `json:"status"` //暂未用到
	Name                    string `json:"name"`
	CreationTimestamp       string `json:"creationTimestamp"`
	Action                  string `json:"action"`
	Count                   int32  `json:"count"`
	Messages                string `json:"short_message"`
	Type                    string `json:"type"`
	FirstTimestamp          string `json:"firstTimestamp"`
	LastTimestamp           string `json:"lastTimestamp"`
	ServiceName             string `json:"serviceName"`
	ResourceVersion         string `json:"resourceVersion"`
	UpdateContent           string
	InvolvedName            string `json:"involvedName"` //add for event's InvolvedObject
	InvolvedNamespace       string `json:"involvedNamespace"`
	InvolvedKind            string `json:"involvedKind"`
	InvolvedResourceVersion string `json:"involvedResourceVersion"`
}

const (
	// CreateEvent event associated with new objects in an informer
	CreateEvent = "CREATE"
	// UpdateEvent event associated with an object update in an informer
	UpdateEvent = "UPDATE"
	// DeleteEvent event associated when an object is removed from an informer
	DeleteEvent = "DELETE"
	// ConfigurationEvent event associated when a configuration object is created or updated
	//ConfigurationEvent = "CONFIGURATION"
)

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
}

/*
New create new diy K8satchEvent
涵盖几乎所有k8s资源对象
handler各自调用，改为统一New之后传给handler 避免重复New
*/
func New(obj interface{}, action string) Event {
	//zlog.Debugf("New got obj,type:%s", reflect.TypeOf(obj))
	var kind, host string
	var kbEvent Event
	objectTypeMeta := utils.GetObjectTypeMetaData(obj)
	kind = objectTypeMeta.Kind
	objectMeta := utils.GetObjectMetaData(obj)
	kbEvent.Namespace = objectMeta.Namespace
	kbEvent.Name = objectMeta.Name
	kbEvent.Action = action
	kbEvent.ResourceVersion = objectMeta.ResourceVersion
	//zlog.Debugf("objectMeta:%+v", objectMeta)
	kbEvent.CreationTimestamp = time.Unix(objectMeta.CreationTimestamp.Unix(), 0).Format("2006-01-02 15:04:05")

	switch object := obj.(type) {
	case *ext_v1beta1.DaemonSet:
		kind = "daemonsets"
		//	case *apps_v1beta1.Deployment:
	case *ext_v1beta1.Deployment:
		kind = "deployments"
		/*if len(object.Spec.Template.Spec.Containers) > 0 {
			s := object.Spec.Template.Spec.Containers[0].Image
		}*/
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *batch_v1.Job:
		kind = "jobs"
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *api_v1.Namespace:
		kind = "namespaces"
	case *ext_v1beta1.Ingress:
		kind = "ingresses"
	case *api_v1.PersistentVolume:
		kind = "persistentvolumes"
	case *api_v1.PersistentVolumeClaim:
		kind = "persistentvolumeclaim"
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *api_v1.Pod:
		kind = "pods"
		host = object.Spec.NodeName
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *api_v1.Node:
		kind = "nodes"
		host = object.ObjectMeta.Name
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastHeartbeatTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *api_v1.ReplicationController:
		kind = "replicationcontrolleres"
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *ext_v1beta1.ReplicaSet:
		kind = "replicasets"
		if action == UpdateEvent {
			condLen := len(object.Status.Conditions)
			if condLen != 0 {
				kbEvent.LastTimestamp = time.Unix(object.Status.Conditions[condLen-1].LastTransitionTime.Time.Unix(), 0).Format("2006-01-02 15:04:05")
			}
		}
	case *api_v1.Service:
		kind = "services"
		kbEvent.Component = string(object.Spec.Type)
	case *api_v1.Secret:
		kind = "secrets"
	case *api_v1.ConfigMap:
		kind = "configmaps"

	case *api_v1.Event:
		kind = "events"
		object = obj.(*api_v1.Event)
		kbEvent.InvolvedKind = object.InvolvedObject.Kind
		kbEvent.InvolvedName = object.InvolvedObject.Name
		kbEvent.InvolvedNamespace = object.InvolvedObject.Namespace
		kbEvent.InvolvedResourceVersion = object.InvolvedObject.ResourceVersion
		//zlog.Debugf("object:%+v", object)
		kbEvent.Type = object.Type
		kbEvent.FirstTimestamp = time.Unix(object.FirstTimestamp.Unix(), 0).Format("2006-01-02 15:04:05")
		kbEvent.LastTimestamp = time.Unix(object.LastTimestamp.Unix(), 0).Format("2006-01-02 15:04:05")

		kbEvent.Reason = object.Reason
		//kbEvent.Kind = object.Kind
		kbEvent.Component = object.Source.Component

		serviceNameRe := regexp.MustCompile(`spec.containers{([^}]+)}`)
		serviceNameOri := object.InvolvedObject.FieldPath

		match := serviceNameRe.FindSubmatch([]byte(serviceNameOri))
		if len(match) >= 2 {
			kbEvent.ServiceName = string(match[1])
		} else {
			zlog.Debugf("正则未能匹配到事件中的serviceName Ori:%s  object.InvolvedObject:%+v", serviceNameOri, object.InvolvedObject)
		}
		kbEvent.Count = object.Count
		host = object.Source.Host
		kbEvent.Component = object.Source.Component
		kbEvent.Messages += object.Message

	case *api_v1.Endpoints:
		kind = "endpoints"
	case *rbacV1.Role:
		kind = "roles"
	case *rbacV1.RoleBinding:
		kind = "rolebindings"
	case *rbacV1.ClusterRole:
		kind = "clusterroles"
	case *rbacV1.ClusterRoleBinding:
		kind = "clusterrolebindings"

	case Event:
		//用于delete情况，此时对象已不存在，因此手动拼接基本信息并生成msg即可
		//删除类型的name是metafunc获取到的namespace/name形式的
		kind = object.Kind
		if strings.Contains(object.Name, "/") {
			metavalue := strings.Split(object.Name, "/")
			kbEvent.Namespace = metavalue[0]
			kbEvent.Name = metavalue[1]
		} else {
			kbEvent.Name = object.Name
		}
		//kbEvent.Namespace = object.Namespace

	default:
		zlog.Warnf("unknown obj reflectType:%v ,obj:%v", reflect.TypeOf(object), obj)
	}
	if host == "" {
		host = "Empty"
	}
	kbEvent.Host = host
	kbEvent.Kind = kind
	kbEvent.Messages += fmt.Sprintf("%s", kbEvent.Message())
	return kbEvent
}

/*
Message returns event message in standard format.
included as a part of event packege to enhance code resuablity across handlers.
*/

func (e *Event) Message() (msg string) {
	// using switch over if..else, since the format could vary based on the kind of the object in future.
	switch e.Kind {
	case "namespaces", "nodes":
		msg = fmt.Sprintf(
			"%s:%s has been %sD",
			e.Kind,
			e.Name,
			e.Action,
		)
	case "events":

	default:
		msg = fmt.Sprintf(
			"%s:%s in namespace %s has been %sD",
			e.Kind,
			e.Name,
			e.Namespace,
			e.Action,
		)
	}
	return msg
}
