package watch

import (
	"testing"
	"time"

	fakeclient "k8s.io/client-go/kubernetes/fake"
)

func TestReleaseWatcher(t *testing.T) {
	client := fakeclient.NewSimpleClientset()

	rw := watch.NewReleaseWatcher(client, "configmap")
	informerFactory := k8sinformers.NewSharedInformerFactory(client, time.Duration(time.Minute))
}
