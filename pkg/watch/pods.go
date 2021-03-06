package watch

import (
	"time"

	"github.com/prometheus/common/log"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher struct {
	informerFactory *cache.SharedIndexInformer
	labels          labels.Selector
}

func NewPodWatcher(selector labels.Selector) PodWatcher {
	return *PodWatcher{
		informerFactory: informers.NewSharedInformerFactory(client, defaultResync),
		labels:          labels,
	}
}

func (pw *PodWatcher) Run() {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: filterPods,
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

func filterPods(obj interface{}) bool {
	pod := obj.(*v1core.Pod)
	return selector.Matches(labels.Set(pod.Labels))
}

func podAdded(obj interface{}) {
	pod := obj.(*v1core.Pod)
	log.Info("added pod ", pod.Name)
}

func podUpdated(oldObj, newObj interface{}) {
	_ = oldObj.(*v1core.Pod)
	newPod := newObj.(*v1core.Pod)
	log.Info("updated pod ", newPod.Name)
	log.Info("pod events")
	log.Info(newPod.Status)
	log.Info("---")
}

func podDeleted(obj interface{}) {
	pod := obj.(*v1core.Pod)
	log.Info("deleted pod ", pod.Name)
}
