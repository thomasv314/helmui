module github.com/thomasv314/helmui

go 1.15

require (
	github.com/deislabs/oras v0.10.1-0.20210302005037-faf2f4b1daa2 // indirect
	github.com/prometheus/common v0.10.0
	github.com/rs/zerolog v1.20.0
	golang.org/x/tools v0.1.0 // indirect
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/cli-runtime v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/klog v1.0.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
