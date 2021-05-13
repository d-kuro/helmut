package helmut_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/d-kuro/helmut"
	"github.com/d-kuro/helmut/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestRenderTemplates(t *testing.T) {
	t.Parallel()

	const (
		releaseName = "foo"
		chartName   = "test-chart"
	)

	tests := []struct {
		name          string
		assertOptions []assert.Option
		wantObject    runtime.Object
	}{
		{
			name: "contains service account",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			wantObject: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", releaseName, chartName),
				},
			},
		},
		{
			name: "contains service",
			assertOptions: []assert.Option{
				assert.WithIgnoreLabels([]string{
					"app.kubernetes.io/managed-by",
					"app.kubernetes.io/version",
					"helm.sh/chart",
				}),
			},
			wantObject: &corev1.Service{
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
							Port:       80,
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
			t.Parallel()

			assert.Contains(t, manifests, tt.wantObject, tt.assertOptions...)
		})
	}
}
