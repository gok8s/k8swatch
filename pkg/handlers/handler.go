package handlers

import (
	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/event"
	"github.com/gok8s/k8swatch/pkg/handlers/alert"
	"github.com/gok8s/k8swatch/pkg/handlers/influxdb"
	"github.com/gok8s/k8swatch/pkg/handlers/rabbitmq"
	"github.com/gok8s/k8swatch/pkg/handlers/webhook"
)

// Handlers is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Init(c config.Config) error
	ObjectCreated(obj event.Event)
	ObjectDeleted(obj event.Event)
	ObjectUpdated(obj event.Event)
}

// Map maps each event handler function to a name for easily lookup
var Map = map[string]interface{}{
	"default":  &Default{},
	"rabbitmq": &rabbitmq.RabbitMq{},
	"influxdb": &influxdb.InfluxDB{},
	"alert":    &alert.Alert{},
	"webhook":  &webhook.Webhook{},
}

// Default handler implements Handlers interface,
// print each event with JSON format
type Default struct {
}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c config.Config) error {
	return nil
}

func (d *Default) ObjectCreated(obj event.Event) {

}

func (d *Default) ObjectDeleted(obj event.Event) {

}

func (d *Default) ObjectUpdated(event.Event) {

}
