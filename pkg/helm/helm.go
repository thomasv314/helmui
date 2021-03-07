package helm

import (
	"log"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

var (
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
)

func Init(helmDriver string) {
	klog.V(6).Infof("init helm client")

	settings = cli.New()
	actionConfig = new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), "", helmDriver, log.Printf); err != nil {
		os.Exit(1)
		klog.Error(err)
	}
}

func GetRelease(name string) (release *release.Release, err error) {
	status := action.NewStatus(actionConfig)
	release, err = status.Run(name)
	return
}

func GetReleaseObjects(name string) (objects []interface{}, err error) {
	release, err := GetRelease(name)

	if err != nil {
		return
	}

	scheme := runtime.NewScheme()
	_ = v1apps.AddToScheme(scheme)
	_ = v1core.AddToScheme(scheme)
	_ = scheme.AllKnownTypes()
	deserializer := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	files := strings.Split(release.Manifest, "---")

	objects = make([]interface{}, 0)

	for i := range files {
		obj, _, decodeErr := deserializer.Decode([]byte(files[i]), nil, nil)

		if decodeErr != nil {
			klog.V(6).Infof("decode error: %s", decodeErr)
			continue
		}

		objects = append(objects, obj)
	}

	return

}
