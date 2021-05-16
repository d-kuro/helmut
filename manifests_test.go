package helmut_test

import (
	"testing"

	"github.com/d-kuro/helmut"
	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStoreAndLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want runtime.Object
	}{
		{
			name: "deployment",
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
		{
			name: "same as deployment",
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
		{
			name: "service",
			want: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
	}

	manifests := helmut.NewManifests()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key, err := helmut.NewObjectKeyFromObject(tt.want)
			if err != nil {
				t.Fatalf("failed to create objectkey: %s", err)
			}

			manifests.Store(key, tt.want)

			got, ok := manifests.Load(key)
			if !ok {
				t.Fatalf("not found object: %s", key)
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("object mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		object runtime.Object
	}{
		{
			name: "deployment",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
		{
			name: "service",
			object: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
		},
	}

	manifests := helmut.NewManifests()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key, err := helmut.NewObjectKeyFromObject(tt.object)
			if err != nil {
				t.Fatalf("failed to create objectkey: %s", err)
			}

			manifests.Store(key, tt.object)

			if _, ok := manifests.Load(key); !ok {
				t.Fatalf("not found object: %s", key)
			}

			manifests.Delete(key)

			if _, ok := manifests.Load(key); ok {
				t.Fatalf("object has not been deleted: %s", key)
			}
		})
	}
}

func TestLength(t *testing.T) {
	t.Parallel()

	object := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	key, err := helmut.NewObjectKeyFromObject(object)
	if err != nil {
		t.Fatalf("failed to create objectkey: %s", err)
	}

	manifests := helmut.NewManifests()

	if manifests.Length() != 0 {
		t.Fatalf("length: got %d, want %d", 0, manifests.Length())
	}

	manifests.Store(key, object)

	if manifests.Length() != 1 {
		t.Fatalf("length: got %d, want %d", 1, manifests.Length())
	}

	manifests.Store(key, object)

	if manifests.Length() != 1 {
		t.Fatalf("length: got %d, want %d", 1, manifests.Length())
	}

	manifests.Delete(key)

	if manifests.Length() != 0 {
		t.Fatalf("length: got %d, want %d", 0, manifests.Length())
	}
}
