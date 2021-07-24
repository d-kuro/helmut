module github.com/d-kuro/helmut

go 1.16

require (
	github.com/google/go-cmp v0.5.6
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/yaml v1.2.0
)

// ref: https://github.com/helm/helm/blob/v3.6.3/go.mod#L50
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
