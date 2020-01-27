module github.com/gok8s/k8swatch

require (
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.1
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/influxdata/influxdb v1.7.8
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olivere/elastic v6.2.25+incompatible
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/xiaomeng79/go-log v2.0.4+incompatible // indirect
	go.uber.org/zap v1.10.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.0.0-20191005115622-2e41325d9e4b
	k8s.io/apimachinery v0.0.0-20191005115455-e71eb83a557c

	k8s.io/client-go v0.0.0-20191005115821-b1fd78950135
	k8s.io/klog v1.0.0
	k8s.io/sample-controller v0.0.0-20191005120943-ac9726f261cc
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect

)

replace (
	github.com/Sirupsen/logrus v1.0.5 => github.com/sirupsen/logrus v1.0.5
	github.com/Sirupsen/logrus v1.3.0 => github.com/Sirupsen/logrus v1.0.6
	github.com/Sirupsen/logrus v1.4.0 => github.com/sirupsen/logrus v1.0.6

)

go 1.13
