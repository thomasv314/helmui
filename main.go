package main

import (
	"flag"
	_ "fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	"github.com/thomasv314/helmui/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	var config *rest.Config
	var err error

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Starting helmui")

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

	helm.Init()
	watch.Init(config)

	stopCh := make(chan struct{})
	releaseWatcher := watch.NewReleaseWatcher()
	releaseWatcher.Run(stopCh)
	for {
		time.Sleep(time.Second)
	}
}
