package elasticsearch

import (
	"context"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/gok8s/k8swatch/utils/zlog"
	"testing"

	"github.com/gok8s/k8swatch/pkg/event"
)

func TestSaver(t *testing.T) {
	item1 := event.Event{
		Namespace:         "cre-abc-def-hig",
		Kind:              "pod",
		Component:         "kubelet",
		Host:              "1",
		Reason:            "",
		Status:            "",
		Name:              "crename-1-haha-2.33",
		CreationTimestamp: "",
		Action:            "",
		Count:             0,
		Messages:          "",
		Type:              "",
		FirstTimestamp:    "",
		LastTimestamp:     "",
		ServiceName:       "",
	}

	client, err := elastic.NewClient(
		elastic.SetSniff(false))

	if err != nil {
		panic(err)
	}
	err = Save(client, "k8swatch", item1)
	if err != nil {
		panic(err)
	}

	resp, err := client.Get().
		Index("k8swatch").
		Id("1").
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	//var actual event.Event
	//err = json.Unmarshal([]byte(resp.Source), &actual)

	fmt.Printf("%v", resp.Source)
	// Verify result
	/*
		if actual != expected.Payload {
			t.Errorf("got %v; expected %v",
				actual, expected)
		}
	*/
}

func TestSearch(t *testing.T) {
	client, err := elastic.NewClient(
		elastic.SetURL([]string{"http://10.10.9.9:9200", "http://10.10.99.209:9200", "http://10.10.96.113:9200"}...),
		elastic.SetSniff(false))

	if err != nil {
		panic(err)
	} //"t-k8s-node-117-90"
	items, e := Search(client, "k8swatch", "", "nodes", "t-k8s-node-52-183", 10)
	//	items, e := Search(client, "k8swatch", "fpdev-settlement", "pods", "finup-laserpay-optimization-3988354325-9cj5q", 1000)
	if e != nil {
		zlog.Errorf("err:%s", e)
	}
	for _, item := range items {
		fmt.Printf("item:%+v\n", item)
	}
}
