package watch

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/helm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	defaultResync time.Duration = 0
	client        *kubernetes.Clientset
)

func Init(config *rest.Config) {
	log.Debug().Msg("Init kubernetes client")

	var err error
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	helm.Init()
}
