package watch

import (
	v1apps "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/klog"
)

func Deployment(deployment *v1apps.Deployment) {
	coreInformers := informers.NewSharedInformerFactory(client, defaultResync)
	informer := coreInformers.Core().V1().Secrets().Informer()

	klog.Infof("deployment detected: %s", deployment.Name)
	klog.Infof("got an informer", informer)
}
