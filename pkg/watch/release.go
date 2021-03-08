package watch

import (
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Release struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (r *Release) Name() string {
	return r.Labels["name"]
}

func (r *Release) Status() string {
	return r.Labels["status"]
}

func (r *Release) Version() int {
	version, _ := strconv.Atoi(r.Labels["version"])
	return version
}

func (r *Release) Logger() zerolog.Logger {
	logger := log.With().
		Str("release-name", r.Name()).
		Str("release-status", r.Status()).
		Int("release-version", r.Version()).
		Logger()
	return logger
}

func ReleaseFromObject(storeType, obj interface{}) *Release {
	var release Release

	if storeType == ConfigMapStoreType {
		cm := obj.(*v1core.ConfigMap)
		release = Release{
			ObjectMeta: cm.ObjectMeta,
			TypeMeta:   cm.TypeMeta,
		}
	} else {
		sec := obj.(*v1core.Secret)
		release = Release{
			ObjectMeta: sec.ObjectMeta,
			TypeMeta:   sec.TypeMeta,
		}
	}

	return &release
}
