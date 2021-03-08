package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	"github.com/thomasv314/helmui/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig *string
	err        error
	config     *rest.Config
	client     *kubernetes.Clientset
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	debug := flag.Bool("debug", false, "sets log level to debug")
	// Init the Kubeconfig
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Info().Msg("Setting debug")
	}

	if err != nil {
		panic(err.Error())
	}

	// Init a k8s + helm client
	// TODO: helm client doesnt currently init from kubeconfig in this code block, uses default
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	helmDriver := os.Getenv("HELM_DRIVER")

	var chosenDriver string
	if helmDriver == "" {
		chosenDriver = watch.SecretStoreType
	} else {
		if helmDriver == watch.SecretStoreType || helmDriver == watch.ConfigMapStoreType {
			chosenDriver = helmDriver
		} else {
			panic(fmt.Errorf("helm driver not supported: %s", helmDriver))
		}
	}

	log.Info().Str("release-driver", chosenDriver).Msg("Starting helmui")

	helm.Init(chosenDriver)

	stopCh := make(chan struct{})
	releaseWatcher := watch.NewReleaseWatcher(client, chosenDriver)
	releaseWatcher.Run(stopCh)
	for {
		time.Sleep(time.Second)
	}
}
