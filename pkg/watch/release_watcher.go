package watch

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
)

type ReleaseWatcher struct {
	Releases map[string]*Release

	client      *kubernetes.Clientset
	storeType   string
	podWatchers PodWatcher

	informerFactory informers.SharedInformerFactory
	informer        cache.SharedIndexInformer
}

type ReleaseMeta struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

const (
	ConfigMapStoreType = "configmap"
	SecretStoreType    = "secret"
)

// Takes a string of either "objs" or "configmap"
func NewReleaseWatcher(client *kubernetes.Clientset, storeType string) *ReleaseWatcher {
	factory := informers.NewSharedInformerFactory(client, DefaultResync)

	var informer cache.SharedIndexInformer
	if storeType == ConfigMapStoreType {
		informer = factory.Core().V1().ConfigMaps().Informer()
	} else {
		informer = factory.Core().V1().Secrets().Informer()
	}

	rw := &ReleaseWatcher{
		client:          client,
		Releases:        make(map[string]*Release),
		storeType:       storeType,
		informerFactory: factory,
		informer:        informer,
	}

	return rw
}

func (rw *ReleaseWatcher) Run(stop chan struct{}) (err error) {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: rw.filterRelease,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    rw.releaseAdded,
			UpdateFunc: rw.releaseUpdated,
			DeleteFunc: rw.releaseDeleted,
		},
	}
	rw.informer.AddEventHandler(&freh)

	log.Info().Msg("Watching for helm releases")

	stop = make(chan struct{})
	rw.informerFactory.Start(stop)
	if !cache.WaitForCacheSync(stop, rw.informer.HasSynced) {
		return fmt.Errorf("WaitForCacheSync failed")
	}

	return
}

func (rw *ReleaseWatcher) releaseAdded(obj interface{}) {
	var release *Release
	currentRelease := ReleaseFromObject(rw.storeType, obj)
	if _, found := rw.Releases[currentRelease.Name()]; found {
		release = rw.Releases[currentRelease.Name()]
	} else {
		release = currentRelease
	}

	sublogger := release.Logger()

	sublogger.Debug().Msg("ReleaseAdded")

	if release.Status() == "pending-upgrade" {
		objects, err := helm.GetReleaseObjects(release.Name())
		if err != nil {
			sublogger.Error().Err(err).Msg("Error getting release objects")
		} else {
			for _, obj := range objects {
				switch obj.(type) {
				case *v1apps.Deployment:
					deployment := obj.(*v1apps.Deployment)
					log.Info().Str("name", deployment.Name).Msg("Deployment detected")
					podSelector, _ := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
					pw := NewPodWatcher(rw.client, podSelector)
					go pw.Run()
				case *v1core.Service:
					service := obj.(*v1core.Service)
					log.Info().Str("name", service.Name).Msg("Service detected")
				}
			}
		}
	}
}

func (rw *ReleaseWatcher) releaseUpdated(oldObj, obj interface{}) {
	var release *Release
	currentRelease := ReleaseFromObject(rw.storeType, obj)
	if _, found := rw.Releases[currentRelease.Name()]; found {
		release = rw.Releases[currentRelease.Name()]
	} else {
		release = currentRelease
	}

	sublogger := release.Logger()
	sublogger.Debug().Msg("Release updated")
}

func (rw *ReleaseWatcher) releaseDeleted(obj interface{}) {
	var release *Release
	currentRelease := ReleaseFromObject(rw.storeType, obj)
	if _, found := rw.Releases[currentRelease.Name()]; found {
		release = rw.Releases[currentRelease.Name()]
	} else {
		release = currentRelease
	}

	sublogger := release.Logger()
	sublogger.Debug().Msg("Release deleted")
}

func (rw *ReleaseWatcher) filterRelease(obj interface{}) bool {
	currentRelease := ReleaseFromObject(rw.storeType, obj)

	for _, mf := range currentRelease.ManagedFields {
		if mf.Manager == "helm" {
			return true
		}
	}

	return false
}
