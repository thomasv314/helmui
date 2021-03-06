package watch

import (
	"time"

	"github.com/thomasv314/helmui/pkg/helm"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type ReleaseWatcher struct {
	informerFactory informers.SharedInformerFactory
}

func NewReleaseWatcher() *ReleaseWatcher {
	return &ReleaseWatcher{
		informerFactory: informers.NewSharedInformerFactory(client, defaultResync),
	}
}

func (rw *ReleaseWatcher) Run() {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: helmSecretFilter,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    helmSecretAdded,
			UpdateFunc: helmSecretUpdated,
			DeleteFunc: helmSecretDeleted,
		},
	}

	klog.Info("Watching for releases (k8s secrets)")
	informer := rw.informerFactory.Core().V1().Secrets().Informer()
	informer.AddEventHandler(&freh)
	stop := make(chan struct{})
	go informer.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func helmSecretAdded(obj interface{}) {
	secret := obj.(*v1core.Secret)
	klog.V(2).Infof("added secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])

	if secret.Labels["status"] == "pending-upgrade" {
		klog.Infof("new deploy detected status=%s release=%s", secret.Name, secret.Labels["version"])
		//					version, _ := strconv.Atoi(secret.Labels["version"])

		objects, err := helm.GetReleaseObjects(secret.Labels["name"])

		if err != nil {
			klog.Errorf("Error getting release objects: %e", err)
		} else {
			for _, obj := range objects {
				switch obj.(type) {
				case *v1apps.Deployment:
					deployment := obj.(*v1apps.Deployment)
					klog.Infof("deployment/%s selector:%s", deployment.Name, deployment.Spec.Selector.MatchLabels)
					podSelector, _ := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
					WatchPods(podSelector)
				case *v1core.Service:
					service := obj.(*v1core.Service)
					klog.Infof("service/%s", service.Name)
				}
			}
		}
	}
}

func helmSecretUpdated(oldObj, obj interface{}) {
	secret := obj.(*v1core.Secret)
	klog.V(2).Infof("updated secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])
}

func helmSecretDeleted(obj interface{}) {
	secret := obj.(*v1core.Secret)
	klog.V(2).Infof("deleted secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])
}

func helmSecretFilter(obj interface{}) bool {
	secret := obj.(*v1core.Secret)
	return secret.Type == "helm.sh/release.v1"
}
