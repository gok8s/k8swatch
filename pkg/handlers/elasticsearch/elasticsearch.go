package elasticsearch

import (
	"context"
	"github.com/olivere/elastic" //current v6
	"github.com/gok8s/k8swatch/utils"

	//	"github.com/olivere/elastic/v7"

	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/event"
	"github.com/gok8s/k8swatch/utils/zlog"
	"go.uber.org/zap"
	"reflect"
	"strings"
	//"gopkg.in/olivere/elastic.v5"
)

type ElasticClt struct {
	Client *elastic.Client
	Conf   config.ElasticsearchConf
}

func (e *ElasticClt) Init(c config.Config) (err error) {
	e.Conf = c.Handlers.Elasticsearch
	e.Client, err = elastic.NewClient(
		elastic.SetURL(e.Conf.Servers...),
		// Must turn off sniff in docker
		elastic.SetSniff(false),
	)
	if err != nil {
		zlog.Error("es Client init error", zap.Error(err))
		return err
	}

	//TODO 通常索引自动创建，此处不需要手动做mapping并确认、创建，
	/*exists, err := e.Client.IndexExists(e.Conf.Index).Do(context.Background())
	if err != nil {
		zlog.Error("es index exists error", zap.Error(err))
		return err
	}
	if !exists {
		_, err := e.Client.CreateIndex(e.Conf.Index).BodyString().Do(context.Background())
		if err != nil {
			panic(err)
		}
	}*/

	return nil
}

func (e *ElasticClt) ObjectCreated(obj event.Event) {
	utils.Retry(func() error {
		err := Save(e.Client, e.Conf.Index, obj)
		return err
	}, "Created 写入ES数据", 5, 10)
}

func (e *ElasticClt) ObjectDeleted(obj event.Event) {
	//do nothing
}

func (e *ElasticClt) ObjectUpdated(obj event.Event) {
	utils.Retry(func() error {
		err := Save(e.Client, e.Conf.Index, obj)
		return err
	}, "Updated 写入ES数据", 5, 10)
}

func Save(client *elastic.Client, index string, item event.Event) (err error) {
	indexService, err := client.Index().
		Index(index).
		Type("k8s2").
		//Id("1").
		BodyJson(item).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		zlog.Error("写入es失败", zap.Error(err))
		return err
	}
	zlog.Infof("Indexed with id=%v, type=%s\n", indexService.Id, indexService.Type)

	return nil
}

func (e *ElasticClt) GetEvents(namespace, kind, kindName string) {

}

func rewriteQueryString(q string) string {
	new := strings.Replace(q, "-", "\\-", -1)
	return new
}

// Search 根据条件查询ES中的event，此处用于查询pod事件，通常参数为namespace，podname,limit
func Search(client *elastic.Client, index, InvolvedNamespace, InvolvedKind, InvolvedName string, limit int) (items []event.Event, err error) {
	zlog.Infof("essearch involvedNamespace:%s involvedKind:%s InvolvedName:%s", InvolvedNamespace, InvolvedKind, InvolvedName)

	namespaceQeury := elastic.NewMatchQuery("involvedNamespace", InvolvedNamespace).Operator("AND")
	kindQuery := elastic.NewMatchQuery("involvedKind", InvolvedKind).Operator("AND")
	nameQuery := elastic.NewMatchQuery("involvedName", InvolvedName).Operator("AND")
	//rangeQuery := elastic.NewRangeQuery("timestamp").From(time.Now().Add(-70 * 24 * time.Hour)).To(time.Now())
	//TODO daterangequery diy

	//TODO 精确查找 https://cloud.tencent.com/developer/article/1077587
	//elastic.NewConstantScoreQuery(namespaceQeury)

	/*queryStr := rewriteQueryString(fmt.Sprintf("namespace:%s AND name:%s AND kind:%s", namespace, name, kind))
	zlog.Debug(queryStr)
	strQuery := elastic.NewQueryStringQuery(queryStr).Escape(true)
	boolQ := elastic.NewBoolQuery()
	boolQ.Must(strQuery)*/

	boolQ := elastic.NewBoolQuery()
	boolQ.Must(namespaceQeury, kindQuery, nameQuery)

	/*boolQ.Filter(namespaceQeury, kindQuery)
	boolQ.Filter(rangeQuery).Boost(1.0)*/

	searchResult, err := client.Search().
		Index(index).
		Query(boolQ).
		//Query(rangeQuery).
		//Sort("id", true).
		From(0).Size(limit).
		Pretty(true).Do(context.Background())

	if err != nil {
		zlog.Error("es查询失败", zap.Error(err))
		return nil, err
	}
	total := searchResult.TotalHits()
	zlog.Infof("Found %d subjects\n", total)
	if total > 0 {
		//item := searchResult.Each(reflect.TypeOf(event.Event{}))
		for _, item := range searchResult.Each(reflect.TypeOf(event.Event{})) {
			if t, ok := item.(event.Event); ok {
				items = append(items, t)
			}
		}
	}
	return items, nil
}
