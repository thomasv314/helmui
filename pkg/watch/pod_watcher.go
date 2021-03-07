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

	Pods map[string]Pod
}

func NewPodWatcher(selector labels.Selector) *PodWatcher {
	factory := informers.NewSharedInformerFactory(client, defaultResync)
	informer := factory.Core().V1().Pods().Informer()

	pw := &PodWatcher{
		informer:        informer,
		informerFactory: factory,
		labels:          selector,
		Pods:            make(map[string]Pod, 0),
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
	k8spod := obj.(*v1core.Pod)
	var pod Pod
	if pod, found := pw.Pods[k8spod.Name]; !found {
		pod = Pod{
			Pod: *k8spod,
		}
		pw.Pods[pod.Name] = pod
	}
	log.Info().Str("pod-name", pod.Name).Str("phase", string(pod.Status.Phase)).Msg("Pod added")
}

func (pw *PodWatcher) podUpdated(oldObj, newObj interface{}) {
	_ = oldObj.(*v1core.Pod)
	k8spod := newObj.(*v1core.Pod)

	var pod Pod
	if pod, found := pw.Pods[k8spod.Name]; !found {
		pod = Pod{
			Pod: *k8spod,
		}
		pw.Pods[pod.Name] = pod
	}

	log.Info().
		Str("pod-name", pod.Name).
		Str("reason", pod.Status.Reason).
		Str("phase", string(pod.Status.Phase)).
		Msg("Pod updated")

	for _, cs := range pod.Status.ContainerStatuses {
		state := ""
		if cs.State.Waiting != nil {
			state = "waiting"
		} else if cs.State.Running != nil {
			state = "running"
		} else if cs.State.Terminated != nil {
			state = "terminated"
		}

		log.Info().
			Str("pod-name", pod.Name).
			Str("container", cs.Name).
			Str("status", state).
			Msg("Updated container")

		if state == "terminated" {
			fmt.Printf("failed logs: %s/%s", pod.Name, cs.Name)
			fmt.Println(pod.FailedLogs(cs.Name))
		}
	}
	//	fmt.Println(pod.Status.ContainerStatuses[0])
}

func (pw *PodWatcher) podDeleted(obj interface{}) {
	pod := obj.(*v1core.Pod)
	log.Info().Str("pod-name", pod.Name).Str("phase", string(pod.Status.Phase)).Msg("Pod deleted")
}
