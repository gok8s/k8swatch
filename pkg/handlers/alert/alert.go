package alert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/event"

	"github.com/gok8s/k8swatch/utils/zlog"
)

/*
	//判断事件性质，区分出管理员类别admin和用户类别user
	//1，基于alerttype进行初步区分，匹配adminType的设置为admin
	//2，对于usertype的，排除正常的kill
	//3，属于管理员的ns则设置为admin
	//4，其他不发送报警
*/

type Alert struct {
	EnableAdminAlert    bool
	EnableAppOwnerAlert bool
	AlertSpeaker        string
}

type AlertMsg struct {
	event.Event
	Subject string `json:"subject"`
}

func (a *Alert) Init(c config.Config) error {
	a.EnableAdminAlert = c.Handlers.Alert.EnableAdminAlert
	a.EnableAppOwnerAlert = c.Handlers.Alert.EnableAppOwnerAlert
	a.AlertSpeaker = c.Handlers.Alert.Server
	return nil
}

func (a *Alert) ObjectCreated(obj event.Event) {
	if obj.Kind == "events" {
		a.AlertWorker(obj)
	}
}

func (a *Alert) ObjectDeleted(obj event.Event) {

}

func (a *Alert) ObjectUpdated(obj event.Event) {
	if obj.Kind == "events" {
		a.AlertWorker(obj)
	}
}

const (
	AppOwner = "appowner"
	Normal   = "normal"
	Admin    = "admin"
	Warning  = "warning"
	Unknown  = "unknown"
)

func (a *Alert) AlertWorker(msg event.Event) error {
	parse, e := time.Parse("2006-01-02 15:04:05", msg.LastTimestamp)
	if e != nil {
		zlog.Error("alertworker 解析LastTimestamp失败", zap.Error(e))
	} else if time.Now().Unix()-parse.Unix() > 1800 {
		zlog.Infof("事件时间在30分钟以前不做报警，详情为:%v", msg)
		return nil
	}
	var subject string
	describe, receiverType := ClassifyEvent(msg)

	clusterName := os.Getenv("clusterName")
	if clusterName == "" {
		zlog.Error("cannot get env clusterName")
	}
	subject = fmt.Sprintf("%s %s %s/%s", clusterName, describe, msg.Namespace, msg.ServiceName)

	//日志记录
	zlog.Info("事件告警记录 "+msg.Messages,
		zap.String("subject", subject),
		zap.String("receiverType", receiverType),
		zap.String("describe", describe),
		zap.String("namespace", msg.Namespace),
		zap.String("reason", msg.Reason),
		zap.String("action", msg.Action),
		zap.String("name", msg.Name),
		zap.String("component", msg.Component),
		zap.Int32("count", msg.Count),
		zap.String("eventType", msg.Type), //normal,warning
		zap.String("eventServiceName", msg.ServiceName),
		zap.String("firstTimestamp", msg.FirstTimestamp),
		zap.String("lastTimestamp", msg.LastTimestamp),
		zap.String("eventSourceHost", msg.Host), //因为应用日志是被filebeat收集,会覆盖host字段，因此这里用eventSourceHost来表示
	)

	alertMsg := AlertMsg{msg, subject}
	if receiverType == "admin" {
		if !a.EnableAdminAlert {
			zlog.Infof("未启用admin报警，不调用alert-speaker")
			return nil
		}
	} else if receiverType == "appowner" {
		if !a.EnableAppOwnerAlert {
			zlog.Infof("未启用appowner报警，不调用alert-speaker")
			return nil
		}
	} else {
		zlog.Infof("非admin,appowner类别，不做报警,subject为:%s", subject)
		return nil
	}
	status, respBytes, err := callAlertSpeaker(alertMsg, receiverType, a.AlertSpeaker)
	if status != http.StatusOK || err != nil {
		errmessage := fmt.Sprintf("调用alert-speaker接口失败status:%v err:%v respBytes:%s", status, err, respBytes)
		zlog.Error(errmessage)
		return errors.New(errmessage)
	} else {
		zlog.Info("调用alert-speaker接口成功")
	}
	return nil
}

// ClassifyEvent 根据event/alltypes中的定义对事件进行分级和更友好的描述，对于未定义的reason则用reason作为描述，级别为Warning的未知事件将发给管理员。
func ClassifyEvent(msg event.Event) (describe, receiverType string) {
	var ok bool
	if describe, ok = event.UserAlertReasonType[msg.Reason]; ok {
		if msg.Reason == "Killing" {
			//todo REGEX
			if strings.Contains(msg.Messages, "Container failed liveness probe.. Container will be killed and recreated") {
				describe = event.UserAlertReasonType["UnhealthKilling"]
				receiverType = AppOwner
			} else if strings.Contains(msg.Messages, "FailedPostStartHook") { //todo 待确认
				describe = "容器启动失败:PostStartHook异常"
				receiverType = AppOwner
			} else {
				zlog.Debugf("正常killing，不报警:%s", msg.Messages)
				receiverType = Normal
			}
		} else if _, ok = event.AdminAlertNS[msg.Namespace]; ok {
			receiverType = Admin
		} else {
			receiverType = AppOwner
		}
	} else if describe, ok = event.AdminAlertReasonType[msg.Reason]; ok {
		receiverType = Admin
		zlog.Infof("Admin类型报警触发，Reason: %v, Message: %s", msg.Reason, msg.Messages)
	} else if describe, ok = event.NormalReasonType[msg.Reason]; ok {
		receiverType = Normal
	} else if describe, ok = event.WarnAlertReasonType[msg.Reason]; ok {
		receiverType = Warning
		zlog.Infof("Warning类型事件，Reason: %v, describe:%s Message: %s", msg.Reason, describe, msg.Messages)
	} else {
		describe = msg.Reason
		if msg.Type == "Warning" {
			zlog.Warnf("ClassifyEvent收到未知的Warning级别事件将升级为Admin Reason:%s evt:%+v", msg.Reason, msg)
			receiverType = Admin
		} else if msg.Type == "Normal" {
			zlog.Infof("ClassifyEvent收到未知的%s级别事件 不作处理 Reason:%s evt:%+v", msg.Type, msg.Reason, msg)
			receiverType = Normal
		} else {
			zlog.Warnf("ClassifyEvent收到未知的%s级别事件 不作处理 Reason:%s evt:%+v", msg.Type, msg.Reason, msg)
			receiverType = Unknown
		}
	}
	return describe, receiverType
}

func callAlertSpeaker(msg AlertMsg, receiverType, url string) (int, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recover Triggered: ", r)
		}
	}()

	url += fmt.Sprintf("?receivertype=%s", receiverType)
	client := &http.Client{}
	jsonValue, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		zlog.Error(err.Error())
	}
	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	} else {
		return 503, nil, err
	}

	if err != nil {
		zlog.Error(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Errorf("wrong status %s code:%d", url, resp.StatusCode)
	}
	bodyBytes, ReadErr := ioutil.ReadAll(resp.Body)
	if ReadErr != nil {
		zlog.Error(err.Error())
	}
	return resp.StatusCode, bodyBytes, err
}
