package watch

import (
	"time"

	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func WatchPods(selector labels.Selector) {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*v1core.Pod)

			podLabels := labels.SelectorFromSet(pod.Labels)
			return podLabels.Matches(selector)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1core.Pod)
				klog.Info("added pod", pod.Name)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldPod := oldObj.(*v1core.Pod)
				newPod := newObj.(*v1core.Pod)
				klog.Info("updated pod", newPod.Name)
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1core.Pod)
				klog.Info("deleted pod", pod.Name)
			},
		},
	}

	coreInformers := informers.NewSharedInformerFactory(client, defaultResync)
	informer := coreInformers.Core().V1().Pods().Informer()

	informer.AddEventHandler(&freh)

	stop := make(chan struct{})
	go informer.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
