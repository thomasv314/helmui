package watch

import (
	"log"
	"time"

	"github.com/thomasv314/helmui/pkg/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
)

const (
	DefaultResync = time.Duration(0)
)

var (
	helmClient   *helm.HelmClient
	clientset    *kubernetes.Clientset
	actionConfig *action.Configuration
	driverType   string
)

func Init(helmDriver string) (err error) {
	envSettings := cli.New()
	actionConfig = new(action.Configuration)

	helmClient = helm.NewHelmClient(actionConfig)

	driverType = helmDriver

	err = actionConfig.Init(envSettings.RESTClientGetter(), envSettings.Namespace(), driverType, log.Printf)
	if err != nil {
		return
	}

	cs, err := actionConfig.KubernetesClientSet()
	if err != nil {
		return
	}

	clientset = cs.(*kubernetes.Clientset)

	return
}
