package config

type Config struct {
	Handlers  Handlers   `yaml:"handlers"`
	Resources []Resource `yaml:"resources"`
	Settings  Settings   `yaml:"settings"`
	K8s       K8s        `yaml:"k8s"`
	//ResyncPeriod  time.Duration
	//SyncRateLimit float64
}

type Resource struct {
	Name   string `yaml:"name"`
	Enable bool   `yaml:"enable"`
}

type K8s struct {
	APIServerHost  string
	KubeConfigFile string
	ClusterName    string
}

type Settings struct {
	HttpPort        int
	EnableProfiling bool
	LogLevel        string
	LogFile         string
	LogStdout       bool
	Threadiness     int
}

type Handlers struct {
	RabbitMq      RabbitMqConf      `yaml:"rabbitmq"`
	Influxdb      InfluxdbConf      `yaml:"influxdb"`
	Alert         AlertConf         `yaml:"alert"`
	Webhook       Webhook           `yaml:"webhook"`
	Elasticsearch ElasticsearchConf `yaml:"elasticsearch"`
}

type ElasticsearchConf struct {
	Enable  bool
	Servers []string
	Index   string
	//Type   string
}

type Webhook struct {
	Url string `json:"url"`
}

type RabbitMqConf struct {
	Enable       bool
	Servers      []string
	TopicName    string
	ExchangeType string
	Durable      bool
	RouteKey     string
	UserName     string
	Password     string
	Vhost        string
}

type InfluxdbConf struct {
	Enable   bool
	Server   string
	UserName string
	Password string
	DBName   string
}

type AlertConf struct {
	Enable              bool   `yaml:"enable"`
	EnableAdminAlert    bool   `yaml:"enableAdminAlert"`
	EnableAppOwnerAlert bool   `yaml:"enableAppOwnerAlert"`
	Server              string `yaml:"server"`
}
