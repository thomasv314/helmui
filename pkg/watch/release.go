package watch

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"

	"k8s.io/client-go/tools/cache"
)

type ReleaseWatcher struct {
	Releases map[string]Release

	storeType   string
	podWatchers PodWatcher

	informerFactory informers.SharedInformerFactory
	informer        cache.SharedIndexInformer
}

type ReleaseMeta struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

type Release struct {
	Name    string
	Status  string
	Version string
}

const (
	ConfigMapStoreType = "configmap"
	SecretStoreType    = "secret"
)

// Takes a string of either "objs" or "configmap"
func NewReleaseWatcher(storeType string) *ReleaseWatcher {
	factory := informers.NewSharedInformerFactory(client, defaultResync)

	var informer cache.SharedIndexInformer
	if storeType == ConfigMapStoreType {
		informer = factory.Core().V1().ConfigMaps().Informer()
	} else {
		informer = factory.Core().V1().Secrets().Informer()
	}

	rw := &ReleaseWatcher{
		Releases:        make(map[string]Release),
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

func (rw *ReleaseWatcher) getObjectMeta(obj interface{}) (releaseMeta ReleaseMeta) {
	releaseMeta = ReleaseMeta{}
	if rw.storeType == ConfigMapStoreType {
		cm := obj.(*v1core.ConfigMap)
		return ReleaseMeta{
			ObjectMeta: cm.ObjectMeta,
			TypeMeta:   cm.TypeMeta,
		}
	} else {
		sec := obj.(*v1core.Secret)
		return ReleaseMeta{
			ObjectMeta: sec.ObjectMeta,
			TypeMeta:   sec.TypeMeta,
		}
	}
}

func (rw *ReleaseWatcher) releaseAdded(obj interface{}) {
	meta := rw.getObjectMeta(obj)

	sublogger := log.With().
		Str("name", meta.Name).
		Str("release-name", meta.Labels["name"]).
		Int("release-count", len(rw.Releases)).
		Str("release-status", meta.Labels["status"]).
		Logger()

	sublogger.Debug().
		Str("type", string(meta.TypeMeta.Kind)).
		Msg("Release added")

	if meta.Labels["status"] == "pending-upgrade" {
		if release, found := rw.Releases[meta.Labels["name"]]; !found {
			release = Release{
				Name:    meta.Labels["name"],
				Status:  meta.Labels["status"],
				Version: meta.Labels["version"],
			}

			rw.Releases[release.Name] = release

			sublogger.Info().
				Str("version", meta.Labels["version"]).
				Int("release-count", len(rw.Releases)).
				Msg("Release detected.")
		}

		objects, err := helm.GetReleaseObjects(meta.Labels["name"])

		if err != nil {
			sublogger.Error().Err(err).Msg("Error getting release objects")
		} else {
			for _, obj := range objects {
				switch obj.(type) {
				case *v1apps.Deployment:
					deployment := obj.(*v1apps.Deployment)
					log.Info().Str("name", deployment.Name).Msg("Deployment detected")
					podSelector, _ := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
					pw := NewPodWatcher(podSelector)
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
	meta := rw.getObjectMeta(obj)
	log.Debug().
		Str("name", meta.Name).
		Str("version", meta.Labels["version"]).
		Str("status", meta.Labels["status"]).
		Msg("Release updated")
}

func (rw *ReleaseWatcher) releaseDeleted(obj interface{}) {
	meta := rw.getObjectMeta(obj)
	log.Debug().
		Str("name", meta.Name).
		Str("version", meta.Labels["version"]).
		Str("status", meta.Labels["status"]).
		Msg("Release deleted")
}

func (rw *ReleaseWatcher) filterRelease(obj interface{}) bool {
	meta := rw.getObjectMeta(obj)

	for _, mf := range meta.ManagedFields {
		if mf.Manager == "helm" {
			return true
		}
	}

	return false
}
