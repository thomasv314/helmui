package watch

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type PodWatcher struct {
	informerFactory informers.SharedInformerFactory
	informer        cache.SharedIndexInformer
	labels          labels.Selector
}

func NewPodWatcher(selector labels.Selector) *PodWatcher {
	factory := informers.NewSharedInformerFactory(client, defaultResync)
	informer := factory.Core().V1().Pods().Informer()

	pw := &PodWatcher{
		informer:        informer,
		informerFactory: factory,
		labels:          selector,
	}

	return pw
}

func (pw *PodWatcher) Run() {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod := obj.(*v1core.Pod)
			return pw.labels.Matches(labels.Set(pod.Labels))
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    pw.podAdded,
			UpdateFunc: pw.podUpdated,
			DeleteFunc: pw.podDeleted,
		},
	}

	pw.informer.AddEventHandler(&freh)

	stop := make(chan struct{})
	go pw.informer.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func (pw *PodWatcher) podAdded(obj interface{}) {
	pod := obj.(*v1core.Pod)
	log.Info().Str("pod-name", pod.Name).Str("phase", string(pod.Status.Phase)).Msg("Pod added")
}

func (pw *PodWatcher) podUpdated(oldObj, newObj interface{}) {
	_ = oldObj.(*v1core.Pod)
	pod := newObj.(*v1core.Pod)

	log.Debug().
		Str("pod-name", pod.Name).
		Str("reason", pod.Status.Reason).
		Str("phase", string(pod.Status.Phase)).
		Msg("Pod updated")

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated != nil {
			log.Info().
				Str("pod-name", pod.Name).
				Str("container", cs.Name).
				Str("status", "terminated").
				Msg("Detected terminated containers")

			fmt.Println(LogsForPod(pod, cs.Name, &cs.State.Terminated.StartedAt))
		}
	}
}

func (pw *PodWatcher) podDeleted(obj interface{}) {
	pod := obj.(*v1core.Pod)
	log.Info().Str("pod-name", pod.Name).Str("phase", string(pod.Status.Phase)).Msg("Pod deleted")
}
