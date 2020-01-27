package controller

import (
	"fmt"
	"time"

	"github.com/gok8s/k8swatch/pkg/event"
	"github.com/gok8s/k8swatch/pkg/handlers"

	"github.com/gok8s/k8swatch/utils"

	"github.com/gok8s/k8swatch/pkg/config"
	"github.com/gok8s/k8swatch/utils/zlog"
	"go.uber.org/zap"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

var serverStartTime time.Time

type Controller struct {
	resourceType  string
	queue         workqueue.RateLimitingInterface
	informer      cache.SharedIndexInformer
	config        config.Config
	eventHandlers []handlers.Handler
}

type CacheMeta struct {
	Key             string
	Kind            string
	Action          string
	ResourceVersion string
}

func NewResourceController(eventHandlers []handlers.Handler, informer cache.SharedIndexInformer, config config.Config, resourceType string) *Controller {
	c := &Controller{
		resourceType:  resourceType,
		informer:      informer,
		config:        config,
		eventHandlers: eventHandlers,
	}
	//c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resourceType)
	var cacheMeta CacheMeta
	var err error
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cacheMeta.Key, err = cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				zlog.Error("cache.MetaNamespaceKeyFunc(obj) failed", zap.Error(err))
				return
			}
			cacheMeta.Action = event.CreateEvent
			cacheMeta.Kind = resourceType
			c.queue.Add(cacheMeta)
			zlog.Debugf("AddFunc queue.add item: %+v c.queue.Len():%d", cacheMeta, c.queue.Len())
		},
		UpdateFunc: func(old, new interface{}) {
			oldEvent := event.New(old, event.UpdateEvent)
			newEvent := event.New(new, event.UpdateEvent)
			if oldEvent.ResourceVersion == newEvent.ResourceVersion {
				zlog.Warnf("ResourceVersion not change in UpdateEvent:%s/%s", newEvent.Namespace, newEvent.Name)
				return
			}
			/*		if diff := cmp.Diff(old, new); diff != "" {
					zlog.Debugf("%s:%s/%s has been Updated:%s", oldEvent.Kind, oldEvent.Namespace, oldEvent.Name, diff)
					newEvent.UpdateContent = diff
				}*/

			cacheMeta.Key, err = cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				zlog.Error("cache.MetaNamespaceKeyFunc(obj) failed", zap.Error(err))
				return
			}
			cacheMeta.Action = event.UpdateEvent
			cacheMeta.Kind = resourceType
			cacheMeta.ResourceVersion = newEvent.ResourceVersion
			c.queue.Add(cacheMeta)
			if resourceType == "nodes" {
				zlog.Debug("UpdateFunc nodes",
					zap.String("nodeName", newEvent.Name),
					zap.String("lastHbTime", newEvent.LastTimestamp),
					zap.String("cacheMeta.Key", cacheMeta.Key))
			}
			zlog.Debugf("UpdateFunc queue.add item :%+v c.queue.Len():%d", cacheMeta, c.queue.Len())

			//}
		},
		DeleteFunc: func(obj interface{}) {
			//因从只监听event到全部监听，因此下面的类型异常处理可暂不做。
			/*	_, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					zlog.Errorf("Received unexpected object: %v", obj)
					return
				}*/
			cacheMeta.Key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				zlog.Error("cache.DeletionHandlingMetaNamespaceKeyFunc(obj) failed", zap.Error(err))
				return
			}
			cacheMeta.Action = event.DeleteEvent
			cacheMeta.Kind = resourceType
			c.queue.Add(cacheMeta)
			zlog.Debugf("DeleteFunc queue.add item :%+v c.queue.Len():%d", cacheMeta, c.queue.Len())
		},
	})
	return c
}

// Run starts the k8swatch controller
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	serverStartTime = time.Now().Local()
	zlog.Infof("Starting %s controller serverStartTime:%s", c.resourceType, serverStartTime)

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced, c.HasSynced) {
		zlog.Errorf("Timed out waiting for caches to sync,controller type:%s", c.resourceType)
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	zlog.Infof("%s controller synced and ready", c.resourceType)

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	zlog.Infof("Started workers")
	<-stopCh
	zlog.Infof("Stopping %s Controller", c.resourceType)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	cacheMeta, quit := c.queue.Get()
	zlog.Debugf("processNextItem get item :%+v c.queue.Len():%d", cacheMeta, c.queue.Len())
	if quit {
		zlog.Warn("c.queue.Get return quit,return...")
		return false
	}
	//	tmpc := cacheMeta.(CacheMeta)
	defer c.queue.Done(cacheMeta)

	var err error
	go func() {
		err = c.process(cacheMeta.(CacheMeta))
	}()
	c.handleErr(err, cacheMeta)
	return true
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < maxRetries {
		zlog.Errorf("Error processing %s (will retry): %v", key.(CacheMeta).Key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	utilruntime.HandleError(err)
	zlog.Errorf("Dropping key %q out of the queue: %v", key, err)
}

func (c *Controller) process(cacheMeta CacheMeta) error {
	obj, exists, err := c.informer.GetIndexer().GetByKey(cacheMeta.Key)
	if err != nil {
		zlog.Errorf("获取对象:%+v失败", cacheMeta.Key, zap.Error(err))
		return err
	}

	if !exists {
		//deleteEvent
		zlog.Infof("对象:%v 已被删除 详情:%+v", cacheMeta.Key, cacheMeta)
		var tmpEvent event.Event
		tmpEvent.Name = cacheMeta.Key
		tmpEvent.Kind = cacheMeta.Kind
		tmpEvent.Action = cacheMeta.Action

		handerObj := event.New(tmpEvent, cacheMeta.Action)
		for _, eventHandler := range c.eventHandlers {
			eventHandler.ObjectDeleted(handerObj)
		}
	} else {
		switch cacheMeta.Action {
		case event.CreateEvent:
			// compare CreationTimestamp and serverStartTime and alert only on latest events
			// Could be Replaced by using Delta or DeltaFIFO
			objectMeta := utils.GetObjectMetaData(obj)
			var handlerObj event.Event
			timeDuration := objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds()
			zlog.Debugf(" objectMeta.CreationTimestamp:%s timeDuration:%f", objectMeta.CreationTimestamp, timeDuration)
			if timeDuration > 0 {
				handlerObj = event.New(obj, cacheMeta.Action)
				for _, eventHandler := range c.eventHandlers {
					eventHandler.ObjectCreated(handlerObj)
				}
			} else {
				handlerObj = event.New(obj, cacheMeta.Action) //tmp add for test
				zlog.Debugf("old resource info,ignoring...%+v timeDuration:%f objectMeta.CreationTimestamp:%s  serverStartTime:%s",
					handlerObj, timeDuration, objectMeta.CreationTimestamp, serverStartTime)
			}
		case event.UpdateEvent:
			handerObj := event.New(obj, cacheMeta.Action)
			zlog.Debug("Process Update nodes", zap.String("nodeName", handerObj.Name), zap.String("lastHbTime", handerObj.LastTimestamp))
			for _, eventHandler := range c.eventHandlers {
				eventHandler.ObjectUpdated(handerObj)
			}

		case event.DeleteEvent:
			//理论上!exists触发代表对象DeleteEvent，应该不会进入到此处
			//todo 待确认 有流量进来
			/*	kbEvent := event.Event{
				Kind:      newEvent.Kind,
				Name:      newEvent.CacheKey,
				Namespace: newEvent.Namespace,  //? where is the namespace?
			}*/

			zlog.Warnf("obj exists and in event.DeleteEvent case obj:%v", obj)
			var tmpEvent event.Event
			tmpEvent.Name = cacheMeta.Key
			tmpEvent.Kind = cacheMeta.Kind
			tmpEvent.Action = cacheMeta.Action
			for _, eventHandler := range c.eventHandlers {
				eventHandler.ObjectDeleted(tmpEvent)
			}
		default:
			zlog.Errorf("Unknown action:%+v", cacheMeta)
		}
	}
	return nil
}
