# Helmut

![](https://github.com/d-kuro/helmut/workflows/main/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/d-kuro/helmut.svg)](https://pkg.go.dev/github.com/d-kuro/helmut)

Helmut is a testing library for Unit Testing of Helm charts.

This library was inspired by the following project:

* [github.com/Waterdrips/helmunit](https://github.com/Waterdrips/helmunit)
* [How to unit-test your helm charts with Golang – Alistair Hey – Cloud Native Platform Engineer](https://blog.heyal.co.uk/unit-testing-helm-charts/)

## Status

:warning: This project is immature and API stability is not guaranteed.

## Install

:warning: Currently you need `replace` in go.mod to import this library:

```text
require (
	github.com/d-kuro/helmut v0.0.6
)

// ref: https://github.com/helm/helm/blob/v3.5.4/go.mod#L51-L54
replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
```

## Usage

Example tests:

```go
func TestChart(t *testing.T) {
	const (
		releaseName = "foo"
		chartName   = "test-chart"
	)

	tests := []struct {
		name          string
		assertOptions []assert.Option
		want          runtime.Object
	}{
		{
			name: "contains service account",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			want: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", releaseName, chartName),
				},
			},
		},
		{
			name: "contains service",
			assertOptions: []assert.Option{
				assert.WithIgnoreLabelKeys(
					"app.kubernetes.io/managed-by",
					"app.kubernetes.io/version",
					"helm.sh/chart",
				),
			},
			want: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", releaseName, chartName),
					Labels: map[string]string{
						"app.kubernetes.io/instance": releaseName,
						"app.kubernetes.io/name":     chartName,
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{
							Port:       8080,
							TargetPort: intstr.FromString("http"),
							Protocol:   corev1.ProtocolTCP,
							Name:       "http",
						},
					},
					Selector: map[string]string{
						"app.kubernetes.io/name":     chartName,
						"app.kubernetes.io/instance": releaseName,
					},
				},
			},
		},
	}

	r := helmut.New()

	manifests, err := r.RenderTemplates(releaseName, filepath.Join("testdata", chartName))
	if err != nil {
		t.Fatalf("failed to render templates: %s", err)
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, manifests, tt.want, tt.assertOptions...)
		})
	}
}
```

Output:

```text
=== RUN   TestChart
=== RUN   TestChart/contains_service_account
=== RUN   TestChart/contains_service
    template_test.go:163: service/foo-test-chart mismatch (-want +got):
          &v1.Service{
                TypeMeta:   {Kind: "Service", APIVersion: "v1"},
                ObjectMeta: {Name: "foo-test-chart", Labels: {"app.kubernetes.io/instance": "foo", "app.kubernetes.io/name": "test-chart"}},
                Spec: v1.ServiceSpec{
                        Ports: []v1.ServicePort{
                                {
                                        Name:        "http",
                                        Protocol:    "TCP",
                                        AppProtocol: nil,
        -                               Port:        8080,
        +                               Port:        80,
                                        TargetPort:  {Type: 1, StrVal: "http"},
                                        NodePort:    0,
                                },
                        },
                        Selector:  {"app.kubernetes.io/instance": "foo", "app.kubernetes.io/name": "test-chart"},
                        ClusterIP: "",
                        ... // 15 identical fields
                },
                Status: {},
          }
--- FAIL: TestChart (0.06s)
    --- PASS: TestChart/contains_service_account (0.00s)
    --- FAIL: TestChart/contains_service (0.00s)
```

Helmut uses [github.com/google/go-cmp](https://github.com/google/go-cmp) to display object diffs.

## Utils

Helmut provides utility functions for testing in the `util` package.

https://pkg.go.dev/github.com/d-kuro/helmut/util
