package webhook

import (
	"fmt"
	"log"
	"os"

	"github.com/gok8s/k8swatch/pkg/config"

	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gok8s/k8swatch/pkg/event"
)

var webhookErrMsg = `
%s

You need to set Webhook url
using "--url/-u" or using environment variables:

export KW_WEBHOOK_URL=webhook_url

Command line flags will override environment variables

`

// Webhook handler implements handler.Handlers interface,
// Notify event to Webhook channel
type Webhook struct {
	Url string
}

type WebhookMessage struct {
	Text string `json:"text"`
}

// Init prepares Webhook configuration
func (m *Webhook) Init(c config.Config) error {
	url := c.Handlers.Webhook.Url

	if url == "" {
		url = os.Getenv("KW_WEBHOOK_URL")
	}

	m.Url = url

	return checkMissingWebhookVars(m)
}

func (m *Webhook) ObjectCreated(obj event.Event) {
	notifyWebhook(m, obj, "created")
}

func (m *Webhook) ObjectDeleted(obj event.Event) {
	notifyWebhook(m, obj, "deleted")
}

func (m *Webhook) ObjectUpdated(obj event.Event) {
	notifyWebhook(m, obj, "updated")
}

func notifyWebhook(m *Webhook, obj event.Event, action string) {
	//e := kbEvent.New(obj, action)

	webhookMessage := prepareWebhookMessage(obj, m)

	err := postMessage(m.Url, webhookMessage)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to %s at %s ", m.Url, time.Now())
}

func checkMissingWebhookVars(s *Webhook) error {
	if s.Url == "" {
		return fmt.Errorf(webhookErrMsg, "Missing Webhook url")
	}

	return nil
}

func prepareWebhookMessage(e event.Event, m *Webhook) *WebhookMessage {
	return &WebhookMessage{
		e.Message(),
	}

}

func postMessage(url string, webhookMessage *WebhookMessage) error {
	message, err := json.Marshal(webhookMessage)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
