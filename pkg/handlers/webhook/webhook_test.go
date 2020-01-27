package webhook

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gok8s/k8swatch/pkg/config"
)

func TestWebhookInit(t *testing.T) {
	s := &Webhook{}
	expectedError := fmt.Errorf(webhookErrMsg, "Missing Webhook url")

	var Tests = []struct {
		webhook config.Webhook
		err     error
	}{
		{config.Webhook{Url: "foo"}, nil},
		{config.Webhook{}, expectedError},
	}

	for _, tt := range Tests {
		c := config.Config{}
		c.Handlers.Webhook = tt.webhook
		if err := s.Init(c); !reflect.DeepEqual(err, tt.err) {
			t.Fatalf("Init(): %v", err)
		}
	}
}
