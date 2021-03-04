package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"

	v1apps "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

type Deploy struct {
	Name    string
	Version int
}

func (d Deploy) Watch() {
	klog.Infof("deploy=%s version=%d status=starting-watch", d.Name, d.Version)

	settings := cli.New()

	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		os.Exit(1)
		klog.Error(err)
	}

	klog.Infof("deploy=%s version=%d status=init-helm", d.Name, d.Version)

	status := action.NewStatus(actionConfig)
	release, err := status.Run(d.Name)

	if err != nil {
		klog.Error(err)
	}

	klog.Infof("deploy=%s version=%d status=got-status", d.Name, d.Version)

	d.parseManifest(release.Manifest)
}

func (d Deploy) parseManifest(manifest string) {
	scheme := runtime.NewScheme()
	_ = v1apps.AddToScheme(scheme)
	_ = v1core.AddToScheme(scheme)
	_ = scheme.AllKnownTypes()
	deserializer := serializer.NewCodecFactory(scheme).UniversalDeserializer()

	files := strings.Split(manifest, "---")

	for i := range files {
		obj, gkv, err := deserializer.Decode([]byte(files[i]), nil, nil)
		fmt.Println("obj", obj)
		fmt.Println("gkv", gkv)
		fmt.Println("err", err)
	}
}
