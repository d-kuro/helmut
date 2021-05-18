package assert_test

import (
	"testing"

	"github.com/d-kuro/helmut/assert"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestOmitMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []assert.Option
		object  runtime.Object
		want    runtime.Object
	}{
		{
			name: "helm labels",
			options: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			object: newNginxDeployment(withDeploymentLabels(
				map[string]string{
					"helm.sh/chart":                "test-chart-0.1.0",
					"app.kubernetes.io/name":       "test-chart",
					"app.kubernetes.io/instance":   "RELEASE-NAME",
					"app.kubernetes.io/version":    "1.16.0",
					"app.kubernetes.io/managed-by": "Helm",
				},
			)),
			want: newNginxDeployment(),
		},
		{
			name: "labels",
			options: []assert.Option{
				assert.WithIgnoreLabelKeys("foo"),
			},
			object: newNginxService(withServiceLabels(
				map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
			)),
			want: newNginxService(withServiceLabels(
				map[string]string{
					"bar": "bar",
				},
			)),
		},
		{
			name: "annotations",
			options: []assert.Option{
				assert.WithIgnoreAnnotationKeys("foo"),
			},
			object: newNginxService(withServiceAnnotations(
				map[string]string{
					"foo": "foo",
					"bar": "bar",
				},
			)),
			want: newNginxService(withServiceAnnotations(
				map[string]string{
					"bar": "bar",
				},
			)),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opt := &assert.IgnoreOption{}
			opt.SetOptions(tt.options...)

			got := assert.OmitMetadata(tt.object, opt)

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("object mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
