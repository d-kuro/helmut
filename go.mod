module github.com/d-kuro/helmut

go 1.16

require (
	github.com/google/go-cmp v0.5.5
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.20.4
	sigs.k8s.io/yaml v1.2.0
)

// ref: https://github.com/helm/helm/blob/v3.5.4/go.mod#L51-L54
replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
