package assert_test

import (
	"fmt"
	"testing"

	"github.com/d-kuro/helmut"
	"github.com/d-kuro/helmut/assert"
	"github.com/d-kuro/helmut/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

const rawManifests = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    helm.sh/chart: test-chart-0.1.0
    app.kubernetes.io/name: test-chart
    app.kubernetes.io/instance: RELEASE-NAME
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    foo: bar
  labels:
    foo: bar
spec:
  selector:
    app: nginx
  ports:
  - port: 80`

type fakeT struct {
	assert.TestingT

	message string
}

func (fakeT) Helper() {}

func (f *fakeT) Error(args ...interface{}) {
	f.message = fmt.Sprint(args...)
}

func (f *fakeT) Errorf(format string, args ...interface{}) {
	f.message = fmt.Sprintf(format, args...)
}

func TestContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		want          bool
		assertOptions []assert.Option
		object        runtime.Object
	}{
		{
			name: "deployment",
			want: true,
			object: newNginxDeployment(
				withDeploymentLabels(map[string]string{
					"helm.sh/chart":                "test-chart-0.1.0",
					"app.kubernetes.io/name":       "test-chart",
					"app.kubernetes.io/instance":   "RELEASE-NAME",
					"app.kubernetes.io/version":    "1.16.0",
					"app.kubernetes.io/managed-by": "Helm",
				}),
			),
		},
		{
			name: "deployment with ignore helm labels",
			assertOptions: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
			want:   true,
			object: newNginxDeployment(),
		},
		{
			name: "service",
			want: true,
			object: newNginxService(
				withServiceLabels(map[string]string{
					"foo": "bar",
				}),
				withServiceAnnotations(map[string]string{
					"foo": "bar",
				}),
			),
		},
		{
			name: "service with ignore labels and annotations",
			want: true,
			assertOptions: []assert.Option{
				assert.WithIgnoreLabelKeys("foo"),
				assert.WithIgnoreAnnotationKeys("foo"),
			},
			object: newNginxService(),
		},
		{
			name: "deployment with transform option",
			want: true,
			assertOptions: []assert.Option{
				// omit replicas
				assert.WithTransformer(func(object runtime.Object) runtime.Object {
					deploy, ok := object.(*appsv1.Deployment)
					if !ok {
						return object
					}

					deploy.Spec.Replicas = nil

					return deploy
				}),
			},
			object: newNginxDeployment(
				withDeploymentReplicas(pointer.Int32Ptr(5)), // original replicas are "3"
				withDeploymentLabels(map[string]string{
					"helm.sh/chart":                "test-chart-0.1.0",
					"app.kubernetes.io/name":       "test-chart",
					"app.kubernetes.io/instance":   "RELEASE-NAME",
					"app.kubernetes.io/version":    "1.16.0",
					"app.kubernetes.io/managed-by": "Helm",
				}),
			),
		},
		{
			name: "additional name",
			want: true,
			assertOptions: []assert.Option{
				assert.WithAdditionalKeys(func(key helmut.ObjectKey) helmut.ObjectKey {
					key.Name = "nginx"

					return key
				}),
			},
			object: newNginxService(
				withServiceName("additional-name"),
				withServiceLabels(map[string]string{
					"foo": "bar",
				}),
				withServiceAnnotations(map[string]string{
					"foo": "bar",
				}),
			),
		},
		{
			name:   "diffs exists service",
			want:   false,
			object: newNginxService(),
		},
		{
			name: "serviceaccount object not included in the manifest",
			want: false,
			object: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy",
				},
			},
		},
	}

	r := helmut.New()

	manifests, err := r.SplitManifests([]byte(rawManifests))
	if err != nil {
		t.Fatalf("failed to split manifests: %s", err)
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &fakeT{}

			got := assert.Contains(fakeT, manifests, tt.object, tt.assertOptions...)
			if got != tt.want {
				t.Errorf("got %t, want %t, message: %s", got, tt.want, fakeT.message)
			}
		})
	}
}

func TestContainsWithRawManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		want         bool
		rawManifests []byte
	}{
		{
			name:         "contains raw manifests",
			want:         true,
			rawManifests: []byte(rawManifests),
		},
	}

	r := helmut.New()

	manifests, err := r.SplitManifests([]byte(rawManifests))
	if err != nil {
		t.Fatalf("failed to split manifests: %s", err)
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fakeT := &fakeT{}

			split, err := util.SplitManifests(tt.rawManifests)
			if err != nil {
				t.Errorf("failed to split manifests: %s", err)
			}

			for _, data := range split {
				got := assert.ContainsWithRawManifest(fakeT, manifests, data)
				if got != tt.want {
					t.Errorf("got %t, want %t, message: %s", got, tt.want, fakeT.message)
				}
			}
		})
	}
}

type deploymentOption func(*appsv1.Deployment)

func withDeploymentLabels(labels map[string]string) deploymentOption {
	return func(deploy *appsv1.Deployment) {
		deploy.Labels = labels
	}
}

func withDeploymentReplicas(replicas *int32) deploymentOption {
	return func(deploy *appsv1.Deployment) {
		deploy.Spec.Replicas = replicas
	}
}

func newNginxDeployment(options ...deploymentOption) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, o := range options {
		o(deploy)
	}

	return deploy
}

type serviceOption func(*corev1.Service)

func withServiceName(name string) serviceOption {
	return func(svc *corev1.Service) {
		svc.SetName(name)
	}
}

func withServiceLabels(labels map[string]string) serviceOption {
	return func(svc *corev1.Service) {
		svc.Labels = labels
	}
}

func withServiceAnnotations(annotations map[string]string) serviceOption {
	return func(svc *corev1.Service) {
		svc.Annotations = annotations
	}
}

func newNginxService(options ...serviceOption) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
			Selector: map[string]string{
				"app": "nginx",
			},
		},
	}

	for _, o := range options {
		o(svc)
	}

	return svc
}
