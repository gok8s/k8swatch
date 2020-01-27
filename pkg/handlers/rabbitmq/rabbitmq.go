package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/gok8s/k8swatch/utils"

	"go.uber.org/zap"

	"github.com/streadway/amqp"
	"github.com/gok8s/k8swatch/utils/zlog"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/event"
)

type RabbitMq struct {
	conf config.RabbitMqConf
	conn *amqp.Connection
	ch   *amqp.Channel
}

type RabbitMqMsg struct {
	event.Event
	subject string
}

func (r *RabbitMq) Init(c config.Config) (err error) {
	r.conf = c.Handlers.RabbitMq
	err = r.getConnection()
	if err != nil {
		zlog.Errorf("初始化Connection失败，error: %v", err)
		return err
	}
	//zlog.Debugf("rabbitmq.conf:%+v", r.conf)
	return nil
}

func (r *RabbitMq) ObjectCreated(obj event.Event) {
	msgbytes, err := json.Marshal(obj)
	if err != nil {
		zlog.Error("将KBEvent解析为json失败", zap.Error(err))
	}
	r.Publish(msgbytes)
}

func (r *RabbitMq) ObjectUpdated(obj event.Event) {
	msgbytes, err := json.Marshal(obj)
	if err != nil {
		zlog.Error("将KBEvent解析为json失败", zap.Error(err))
	}
	r.Publish(msgbytes)
}

func (r *RabbitMq) ObjectDeleted(obj event.Event) {
	msgbytes, err := json.Marshal(obj)
	if err != nil {
		zlog.Error("将KBEvent解析为json失败", zap.Error(err))
	}
	r.Publish(msgbytes)
}

func (rmc *RabbitMq) Publish(msgBody []byte) {
	if rmc.ch == nil {
		zlog.Info("channel已关闭，将重连...")
		err := rmc.getChannel()
		if err != nil {
			zlog.Errorf("重连Channel失败，无法发送消息,error is: %v", err)
			return
		}
	}
	var msg event.Event
	json.Unmarshal(msgBody, &msg)

	var err error
	utils.Retry(func() error {
		err = rmc.ch.Publish(
			rmc.conf.TopicName, // exchange
			rmc.conf.RouteKey,  // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				ContentType:  "text/plain",
				Body:         msgBody,
				DeliveryMode: 2,
			})
		return err
	}, "发送MQ消息", 5, 10)

	//TODO client重连
	if err != nil {
		zlog.Errorf("发送消息失败,消息为:%s ,错误为： %v", msgBody, err)
		return
	}
	zlog.Info("发送mq消息成功 "+msg.Messages,
		zap.String("namespace", msg.Namespace),
		zap.String("name", msg.Name),
		zap.String("action", msg.Action),
		zap.String("kind", msg.Kind),
	)
}

func (rmc *RabbitMq) getConnection() (err error) {
	var mqhost string
	if len(rmc.conf.Servers) > 1 {
		mqhost = rmc.conf.Servers[rand.IntnRange(0, len(rmc.conf.Servers))]
	} else if len(rmc.conf.Servers) <= 0 {
		zlog.Error("rmc.conf.RabbitMQHosts is empty")
		return fmt.Errorf("rmc.conf.RabbitMQHosts is empty")
	} else {
		mqhost = rmc.conf.Servers[0]
	}
	connstr := fmt.Sprintf("amqp://%s:%s@%s/%s", rmc.conf.UserName, rmc.conf.Password, mqhost, rmc.conf.Vhost)
	for i := 0; i < 3; i++ {
		rmc.conn, err = amqp.Dial(connstr)
		if err != nil {
			zlog.Errorf("不能连接到MQ: %s,会重试3遍，现在是重试第%d遍!! 错误:  %v", connstr, i, err)
			if i == 2 {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func (rmc *RabbitMq) getChannel() (err error) {

	for i := 0; i < 3; i++ {
		rmc.ch, err = rmc.conn.Channel()
		if err != nil {
			zlog.Errorf("初始化channel失败，会重试3遍，现在是重试第%d遍!! 错误:  %v", i, err)
			rmc.getConnection()
			if i == 2 {
				return err
			}
			continue
		}

		_, err = rmc.ch.QueueDeclarePassive(rmc.conf.RouteKey,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			_, err = rmc.ch.QueueDeclare(
				rmc.conf.RouteKey,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				zlog.Errorf("声明queue:%s 失败，会重试3遍，现在是重试第%d遍,错误:%v", rmc.conf.RouteKey, i, err)
				if i == 2 {
					return err
				}
				continue
			} else {
				zlog.Debugf("声明queue:%s成功", rmc.conf.RouteKey)
			}
		}

		err = rmc.ch.ExchangeDeclarePassive(
			rmc.conf.TopicName,    // name
			rmc.conf.ExchangeType, // type
			rmc.conf.Durable,      // durable
			false,                 // auto-deleted
			false,                 // internal
			false,                 // no-wait
			nil,                   // arguments
		)
		if err != nil {
			err = rmc.ch.ExchangeDeclare(
				rmc.conf.TopicName,    // name
				rmc.conf.ExchangeType, // type
				rmc.conf.Durable,      // durable
				false,                 // auto-deleted
				false,                 // internal
				false,                 // no-wait
				nil,                   // arguments
			)
			if err != nil {
				zlog.Errorf("声明Exchange失败:,会重试3遍，现在是重试第%d遍!! 错误:  %v", i, err)
				if i == 2 {
					return err
				}
				continue
			}
		}
		break
	}
	return nil
}

func (rmc *RabbitMq) Close() {
	if rmc.ch != nil {
		rmc.ch.Close()
	}
	if rmc.conn != nil {
		rmc.conn.Close()
	}
}
