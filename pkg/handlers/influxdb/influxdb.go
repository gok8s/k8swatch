package influxdb

import (
	"fmt"
	"github.com/gok8s/k8swatch/pkg/event"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/utils/zlog"
	api_v1 "k8s.io/api/core/v1"
)

type InfluxDB struct {
	addr     string
	username string
	password string
	dbname   string
	cli      client.Client
	bp       client.BatchPoints
}

func (idc *InfluxDB) Init(c config.Config) error {
	idc.addr = c.Handlers.Influxdb.Server
	idc.username = c.Handlers.Influxdb.UserName
	idc.password = c.Handlers.Influxdb.Password
	idc.dbname = c.Handlers.Influxdb.DBName
	idc.cli, idc.bp = idc.getClient()

	if idc.cli == nil || idc.bp == nil {
		err := "InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!"
		zlog.Errorf(err)
		idc.cli, idc.bp = idc.getClient()
		if idc.cli == nil || idc.bp == nil {
			zlog.Errorf("retry failed..." + err)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (idc *InfluxDB) Close() {
	idc.cli.Close()
}

func (idc *InfluxDB) ObjectCreated(obj event.Event) {
	if obj.Kind == "events" {
		idc.RecordEventToInflux(obj)
	}
}

func (idc *InfluxDB) ObjectDeleted(obj event.Event) {

}

func (idc *InfluxDB) ObjectUpdated(obj event.Event) {
	if obj.Kind == "events" {
		idc.RecordEventToInflux(obj)
	}
}

func (idc *InfluxDB) getClient() (client.Client, client.BatchPoints) {
	insecureSkipVerify := false
	if idc.username == "" || idc.password == "" {
		insecureSkipVerify = true
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:               idc.addr,
		Username:           idc.username,
		Password:           idc.password,
		UserAgent:          "",
		Timeout:            0,
		InsecureSkipVerify: insecureSkipVerify,
		TLSConfig:          nil,
		Proxy:              nil,
	})

	if err != nil {
		zlog.Errorf("Create influxdb client error:  %v", err)
		return nil, nil
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  idc.dbname,
		Precision: "s"})

	if err != nil {
		zlog.Errorf("Batch Point return err:  %v", err)
		return nil, nil
	}
	return c, bp
}

/*
Write 执行写入动作 会先判断连接状态
*/
func (idc *InfluxDB) Write(measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	if idc.cli == nil || idc.bp == nil {
		zlog.Errorf("InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!")
		idc.cli, idc.bp = idc.getClient()
		if idc.cli == nil || idc.bp == nil {
			zlog.Errorf("InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!  retry failed")
			return
		}
	}
	pt, err := client.NewPoint(measurement, tags, fields, t)
	if err != nil {
		zlog.Error("InfluxdbWrite NewPoint failed", zap.Error(err))
		return
	}
	idc.bp.AddPoint(pt)
	if err := idc.cli.Write(idc.bp); err != nil {
		zlog.Error("InfluxdbWrite failed", zap.Error(err))
		return
	}
	zlog.Debugf("InfluxdbWrite Successed: %+v", tags)
}

/*
Get 对资源进行查询
*/
func (idc *InfluxDB) Get(cmd string) (res []client.Result, err error) {
	if idc.cli == nil || idc.bp == nil {
		zlog.Errorf("InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!")
		idc.cli, idc.bp = idc.getClient()
		if idc.cli == nil || idc.bp == nil {
			zlog.Errorf("InfluxDBClientlog:Create DB Client failed due to client or bp is null!!!!  retry failed")
			return
		}
	}
	q := client.Query{
		Command:   cmd,
		Database:  idc.dbname,
		Precision: "s",
	}

	if response, err := idc.cli.Query(q); err == nil {
		if response.Error() != nil {
			zlog.Errorf("Response.res  %v", res)
			return res, response.Error()
		}
		res = response.Results
	} else {
		zlog.Errorf("get error:  %v", err)
		return res, err
	}
	return res, nil
}

/*
RecordEventToInflux 将事件object转换为tags和fields并写入influxdb
*/
var x api_v1.Event

func (idc *InfluxDB) RecordEventToInflux(e event.Event) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	measurement := "k8sevents" //todo 可配置
	tags["namespace"] = e.Namespace
	tags["kind"] = e.Kind
	tags["kind_name"] = e.Name //todo k8sevent的involvedObject有name,但event.Event未包含
	tags["reason"] = e.Reason
	tags["type"] = e.Type
	fields["message"] = e.Message
	fields["count"] = e.Count
	fields["source_component"] = e.Component
	fields["source_host"] = e.Host
	fields["firstTimestamp"] = e.FirstTimestamp
	fields["lastTimestamp"] = e.LastTimestamp
	fields["createTimestamp"] = e.CreationTimestamp //todo .UnixNano()
	fields["evt_name"] = e.Name
	rv, _ := strconv.Atoi(e.ResourceVersion)

	fields["resourceVersion"] = e.ResourceVersion
	parse, err := time.Parse("2006-01-02 15:04:05", e.CreationTimestamp)

	if err != nil {
		zlog.Error("RecordEventToInflux parse time error", zap.Error(err))
	}
	strcreated := parse.Unix() + int64(rv)
	t := time.Unix(strcreated, 0)

	zlog.Debugf("InfluxDBClientlog:measurement: %v  ns:  %v; kind:   %v ; Name:   %v; Reason:    %v ;Type:   %v ;"+
		"Message:  %v; Count:    %v;  Component:    %v;  Host:   %v ; FirstTimestamp:   %v, Action: %v ",
		measurement, e.Namespace, e.Kind, e.Name, e.Reason, e.Type,
		e.Message, e.Count, e.Component, e.Host, e.FirstTimestamp, e.Action)

	idc.Write(measurement, tags, fields, t)
}
