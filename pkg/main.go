package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/gok8s/k8swatch/pkg/handlers/elasticsearch"
	"github.com/gok8s/k8swatch/utils/zlog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gok8s/k8swatch/pkg/handlers/alert"

	"github.com/gok8s/k8swatch/pkg/handlers/influxdb"

	"github.com/gok8s/k8swatch/pkg/handlers"
	"github.com/gok8s/k8swatch/pkg/handlers/rabbitmq"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/pkg/controller"

	wapi "github.com/gok8s/k8swatch/pkg/api"
	"github.com/gok8s/k8swatch/utils"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func Start(config config.Config) {
	//设置系统环境变量-clusterName
	err := os.Setenv("clusterName", config.K8s.ClusterName)
	if err != nil {
		zlog.Errorf("为clusterName设置系统环境变量失败,err:%s", err)
	}
	var eventHandlers []handlers.Handler
	if config.Handlers.RabbitMq.Enable {
		zlog.Info("启用rabbitmq handler")
		eventHandler := new(rabbitmq.RabbitMq)
		if err := eventHandler.Init(config); err != nil {
			zlog.Error(err.Error())
		} else {
			defer eventHandler.Close()
			eventHandlers = append(eventHandlers, eventHandler)
		}
	}
	if config.Handlers.Influxdb.Enable {
		zlog.Info("启用influxdb handler")
		eventHandler := new(influxdb.InfluxDB)
		if err := eventHandler.Init(config); err != nil {
			zlog.Error(err.Error())
		} else {
			defer eventHandler.Close()
			eventHandlers = append(eventHandlers, eventHandler)
		}
	}
	if config.Handlers.Alert.Enable {
		zlog.Info("启用alert-speaker handler")
		eventHandler := new(alert.Alert)
		if err := eventHandler.Init(config); err != nil {
			zlog.Error(err.Error())
		} else {
			eventHandlers = append(eventHandlers, eventHandler)
		}
	}

	if config.Handlers.Elasticsearch.Enable {
		zlog.Info("启用elasticsearch handler")
		eventHandler := new(elasticsearch.ElasticClt)
		if err := eventHandler.Init(config); err != nil {
			zlog.Fatal(err.Error())
		}
		eventHandlers = append(eventHandlers, eventHandler)
	}

	utils.Init(config)
	var kubeClient cache.Getter

	stopCh := make(chan struct{})
	defer close(stopCh)

	for _, resource := range config.Resources {
		if !resource.Enable {
			continue
		}
		var ok bool
		if kubeClient, ok = utils.ResourceGetterMap[resource.Name]; !ok {
			zlog.Errorf("ResourceGetterMap 未找到%s对应的restclient", resource.Name)
			continue
		}
		object, ok := utils.RtObjectMap[resource.Name]
		if !ok {
			zlog.Errorf("ResourceGetterMap 未找到%s对应的object类型", resource.Name)
			continue
		}

		lw := cache.NewListWatchFromClient(
			kubeClient,          // 客户端
			resource.Name,       // 被监控资源类型
			"",                  // 被监控命名空间
			fields.Everything()) // 选择器，减少匹配的资源数量
		informer := cache.NewSharedIndexInformer(lw, object, 0, cache.Indexers{})
		c := controller.NewResourceController(eventHandlers, informer, config, resource.Name)

		go c.Run(config.Settings.Threadiness, stopCh)
		zlog.Infof("resource:%s 的控制器已启动", resource.Name)
	}

	mux := http.NewServeMux()
	eapi := wapi.NewQueryApi(config)

	go registerHandlers(eapi, config.Settings.EnableProfiling, config.Settings.HttpPort, mux)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

/*
 * 注册相关的api,profiling
 */
func registerHandlers(eapi wapi.ElasticSearchApi, enableProfiling bool, port int, mux *http.ServeMux) {
	mux.HandleFunc("/events", eapi.GetPodEvt)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal("{status: ok}")
		w.Write(b)
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			zlog.Errorf("unexpected error: %v", err)
		}
	})

	if enableProfiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/heap", pprof.Index)
		mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
		mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
		mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
		mux.HandleFunc("/debug/pprof/block", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	zlog.Fatalf("%s", server.ListenAndServe())
}
