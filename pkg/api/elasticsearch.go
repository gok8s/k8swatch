package api

import (
	"encoding/json"
	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/event"
	"github.com/gok8s/k8swatch/pkg/handlers/elasticsearch"
	"github.com/gok8s/k8swatch/utils/zlog"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ElasticSearchApi struct {
	*elasticsearch.ElasticClt
}

func NewQueryApi(config config.Config) ElasticSearchApi {
	nea := new(elasticsearch.ElasticClt)
	if err := nea.Init(config); err != nil {
		zlog.Error(err.Error())
	}
	return ElasticSearchApi{nea}
}

// GetPodEvt 类似kubectl get event
func (ea *ElasticSearchApi) GetPodEvt(w http.ResponseWriter, r *http.Request) {
	InvolvedKind := "Pod"
	InvolvedNamespace := r.FormValue("namespace")
	InvolvedName := r.FormValue("kind_name")
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit == 0 {
		limit = 10
	}
	//of, _ := strconv.Atoi(r.FormValue("offset"))

	res, err := elasticsearch.Search(ea.Client, ea.Conf.Index, InvolvedNamespace, InvolvedKind, InvolvedName, limit)
	if err != nil {
		zlog.Error("es query error", zap.Error(err))
		json.NewEncoder(w).Encode(err)
		return
	}
	json.NewEncoder(w).Encode(FmtRes(res))
}

/*
FmtRes 临时兼容之前的返回格式
{
"count": 1,
"createTimestamp": 1571382090000000000,
"evt_name": "test-k8s-node-58-172.15ceabafa804c94f",
"firstTimestamp": "2019-10-18 15:01:30",
"kind": "Node",
"kind_name": "test-k8s-node-58-172",
"lastTimestamp": "2019-10-18 15:01:30",
"message": "Node test-k8s-node-58-172 event: Removing Node test-k8s-node-58-172 from Controller",
"namespace": "default",
"reason": "RemovingNode",
"resourceVersion": 0,
"source_component": "node-controller",
"source_host": "",
"time": "2019-10-18 15:01:30",
"type": "Normal"
},
*/

func FmtRes(res []event.Event) (retRes []map[string]interface{}) {
	if len(res) == 0 {
		zlog.Warn("len of res is 0")
		return retRes
	}
	for _, r := range res {
		resMap := make(map[string]interface{})
		resMap["count"] = r.Count
		timeParse := strings.Replace(r.CreationTimestamp, "\u00A0", " ", -1)
		parse, e := time.Parse("2006-01-02 15:04:05", timeParse)
		if e != nil {
			zlog.Errorf("parse time error:%s parseresult:%s ori filed:%s", e, parse, r.CreationTimestamp)
		}
		resMap["createTimestamp"] = parse
		resMap["evt_name"] = r.Name
		resMap["firstTimestamp"] = strings.Replace(r.FirstTimestamp, "\u00A0", " ", -1)
		resMap["kind"] = r.Kind
		resMap["kind_name"] = r.Name
		resMap["lastTimestamp"] = strings.Replace(r.LastTimestamp, "\u00A0", " ", -1)
		resMap["message"] = r.Messages
		resMap["namespace"] = r.Namespace
		resMap["reason"] = r.Reason
		resMap["resourceVersion"] = ""
		resMap["source_component"] = r.Component
		resMap["source_host"] = r.Host
		resMap["time"] = strings.Replace(r.LastTimestamp, "\u00A0", " ", -1)
		resMap["type"] = r.Type
		retRes = append(retRes, resMap)
	}
	return retRes
}

func Extract(res []event.Event) (retRes []map[string]interface{}) {
	if len(res) == 0 {
		zlog.Error("len of res is 0")
		return retRes
	}
	for _, r := range res {
		resMap := make(map[string]interface{})
		t := reflect.TypeOf(r)
		v := reflect.ValueOf(r)
		for i := 0; i < t.NumField(); i++ {
			resMap[t.Field(i).Name] = v.Field(i).Interface()
		}

		retRes = append(retRes, resMap)
	}
	return retRes
}
