package main

import (
	"flag"
	_ "fmt"
	"path/filepath"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

var (
	defaultResync time.Duration = 0
	client        *kubernetes.Clientset
)

func main() {
	var kubeconfig *string
	var config *rest.Config
	var err error

	klog.InitFlags(nil)

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)

	flag.Parse()
	if err != nil {
		panic(err.Error())
	}

	klog.Info("Init client")
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watch()

}

func watch() {
	freh := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			secret := obj.(*v1.Secret)
			return secret.Type == "helm.sh/release.v1"
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.V(2).Infof("added secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])

				if secret.Labels["status"] == "pending-upgrade" {
					klog.Infof("new deploy detected status=%s release=%s", secret.Name, secret.Labels["version"])
					version, _ := strconv.Atoi(secret.Labels["version"])
					deploy := Deploy{
						Name:    secret.Labels["name"],
						Version: version,
					}

					deploy.Watch()
				}
			},
			UpdateFunc: func(oldObj, obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.V(2).Infof("updated secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])
			},
			DeleteFunc: func(obj interface{}) {
				secret := obj.(*v1.Secret)
				klog.V(2).Infof("deleted secret/%s - status=%s release=%s", secret.Name, secret.Labels["status"], secret.Labels["version"])
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
