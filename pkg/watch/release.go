package watch

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"

	"k8s.io/client-go/tools/cache"
)

type ReleaseWatcher struct {
	informerFactory informers.SharedInformerFactory
	informer        cache.SharedIndexInformer
}

func NewReleaseWatcher() *ReleaseWatcher {
	factory := informers.NewSharedInformerFactory(client, defaultResync)
	informer := factory.Core().V1().Secrets().Informer()

	rw := &ReleaseWatcher{
		informerFactory: factory,
		informer:        informer,
	}

	return rw
}

func (rw *ReleaseWatcher) Run(stop chan struct{}) (err error) {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: helmSecretFilter,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    helmSecretAdded,
			UpdateFunc: helmSecretUpdated,
			DeleteFunc: helmSecretDeleted,
		},
	}

	rw.informer.AddEventHandler(&freh)
	log.Info().Msg("Watching for helm releases")
	stop = make(chan struct{})
	rw.informerFactory.Start(stop)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stop, rw.informer.HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return
}

func helmSecretAdded(obj interface{}) {
	secret := obj.(*v1core.Secret)

	sublogger := log.With().
		Str("name", secret.Labels["name"]).
		Logger()

	sublogger.Debug().
		Str("type", string(secret.Type)).
		Msg("Secret added")

	if secret.Labels["status"] == "pending-upgrade" {
		sublogger.Info().
			Str("version", secret.Labels["version"]).
			Msg("Release detected")

		objects, err := helm.GetReleaseObjects(secret.Labels["name"])

		if err != nil {
			sublogger.Error().Err(err).Msg("Error getting release objects")
		} else {
			for _, obj := range objects {
				switch obj.(type) {
				case *v1apps.Deployment:
					deployment := obj.(*v1apps.Deployment)
					log.Info().Str("name", deployment.Name).Msg("Deployment detected")
					//					podSelector, _ := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
					//					pw := NewPodWatcher(podSelector)
					//					pw.Run()
				case *v1core.Service:
					service := obj.(*v1core.Service)
					log.Info().Str("name", service.Name).Msg("Service detected")
				}
			}
		}
	}
}

func helmSecretUpdated(oldObj, obj interface{}) {
	secret := obj.(*v1core.Secret)
	log.Debug().
		Str("name", secret.Name).
		Str("version", secret.Labels["version"]).
		Str("status", secret.Labels["status"]).
		Msg("Release updated")
}

func helmSecretDeleted(obj interface{}) {
	secret := obj.(*v1core.Secret)
	log.Debug().
		Str("name", secret.Name).
		Str("version", secret.Labels["version"]).
		Str("status", secret.Labels["status"]).
		Msg("Release deleted")
}

func helmSecretFilter(obj interface{}) bool {
	secret := obj.(*v1core.Secret)

	if secret.Type == "helm.sh/release.v1" {
		if secret.Labels["status"] == "superseded" {
			return false
		} else {
			return true
		}
	}

	return false
}
