package watch

import (
	"time"

	"github.com/thomasv314/helmui/pkg/helm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var (
	defaultResync time.Duration = 0
	client        *kubernetes.Clientset
)

func Init(config *rest.Config) {
	klog.Info("Init client")

	var err error
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	helm.Init()
}
