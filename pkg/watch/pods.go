package watch

import (
	"time"

	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type PodWatcher struct {
	informerFactory informers.SharedInformerFactory
	labels          labels.Selector
}

func NewPodWatcher(selector labels.Selector) *PodWatcher {
	return &PodWatcher{
		informerFactory: informers.NewSharedInformerFactory(client, defaultResync),
		labels:          selector,
	}
}

func (pw *PodWatcher) Run() {
	klog.Info("Running pod watcher")

	freh := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*v1core.Pod)
			return pw.labels.Matches(labels.Set(pod.Labels))
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    podAdded,
			UpdateFunc: podUpdated,
			DeleteFunc: podDeleted,
		},
	}

	informer := pw.informerFactory.Core().V1().Pods().Informer()
	informer.AddEventHandler(&freh)
	stop := make(chan struct{})
	go informer.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func podAdded(obj interface{}) {
	pod := obj.(*v1core.Pod)
	klog.Infof("pod=%s event=added", pod.Name)
}

func podUpdated(oldObj, newObj interface{}) {
	_ = oldObj.(*v1core.Pod)
	pod := newObj.(*v1core.Pod)
	klog.Infof("pod=%s event=updated msg=%s", pod.Name, pod.Status.Message)
}

func podDeleted(obj interface{}) {
	pod := obj.(*v1core.Pod)
	klog.Info("deleted pod ", pod.Name)
}
