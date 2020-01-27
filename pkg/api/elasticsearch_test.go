package api

import (
	"fmt"
	"github.com/gok8s/k8swatch/pkg/event"
	"testing"
)

func TestExtract(t *testing.T) {
	a := event.Event{
		Namespace:         "1",
		Kind:              "2",
		Component:         "3",
		Host:              "4",
		Reason:            "5",
		Status:            "6",
		Name:              "7",
		CreationTimestamp: "8",
		Action:            "9",
		Count:             0,
		Messages:          "10",
		Type:              "11",
		FirstTimestamp:    "12",
		LastTimestamp:     "13",
		ServiceName:       "14",
	}
	b := event.Event{
		Namespace:         "1",
		Kind:              "2",
		Component:         "3",
		Host:              "4",
		Reason:            "5",
		Status:            "6",
		Name:              "7",
		CreationTimestamp: "8",
		Action:            "9",
		Count:             0,
		Messages:          "10",
		Type:              "11",
		FirstTimestamp:    "12",
		LastTimestamp:     "13",
		ServiceName:       "14",
	}
	c := []event.Event{a, b}
	retRes := Extract(c)
	for _, x := range retRes {
		fmt.Printf("%+v\n", x)
		/*for k, v := range x {
			fmt.Printf("k:%v v:%v\n", k, v)
		}*/
	}
}
