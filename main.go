package main

import (
	"flag"
	_ "fmt"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"
)

var (
	defaultResync time.Duration = 0
)

func main() {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	flag.Parse()
	if err != nil {
		panic(err.Error())
	}

	klog.Info("Init client")
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	freh := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			secret := obj.(*v1.Secret)
			return secret.Type == "helm.sh/release.v1"
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.Info("Added secret/", secret.Name)
			},
			UpdateFunc: func(oldObj, obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.Info("Updated secret/", secret.Name)
			},
			DeleteFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.Info("Deleted secret", secret.Name)
			},
		},
	}

	klog.Info("Start Secret Informers")

	coreInformers := informers.NewSharedInformerFactory(client, defaultResync)
	informer := coreInformers.Core().V1().Secrets().Informer()

	informer.AddEventHandler(&freh)

	stop := make(chan struct{})
	go informer.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
