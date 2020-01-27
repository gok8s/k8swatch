package rabbitmq

import (
	"github.com/gok8s/k8swatch/pkg/event"
	"testing"

	"github.com/gok8s/k8swatch/utils"

	"github.com/spf13/viper"
	"github.com/gok8s/k8swatch/pkg/config"
)

func TestObjectCreated(t *testing.T) {
	utils.InitConfig()
	var config *config.Config
	viper.Unmarshal(&config)

	t.Log("启用rabbitmq handler")

	r := new(RabbitMq)
	if err := r.Init(*config); err != nil {
		t.Error(err)
	}
	obj := event.Event{}
	r.ObjectCreated(obj)
	r.Close()
}
