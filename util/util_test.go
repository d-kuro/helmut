package util_test

import (
	"fmt"
	"testing"

	"github.com/d-kuro/helmut/util"
	"github.com/google/go-cmp/cmp"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

const (
	rawManifests = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
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
spec:
  selector:
    app: nginx
  ports:
  - port: 80`

	deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
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
        - containerPort: 80`

	serviceManifest = `apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
  - port: 80`

	issuerManifest = `apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: test-issuer
spec:
  selfSigned: {}`

	exampleCustomResourceManifest = `apiVersion: example.com/v1
kind: Example
metadata:
  name: test-example
spec:
  example: {}`
)

func TestRawManifestToObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest []byte
		want     *schema.GroupVersionKind
	}{
		{
			name:     "deployment",
			manifest: []byte(deploymentManifest),
			want: &schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
		{
			name:     "service",
			manifest: []byte(serviceManifest),
			want: &schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			},
		},
		{
			name:     "custom resource issuer",
			manifest: []byte(issuerManifest),
			want: &schema.GroupVersionKind{
				Group:   "cert-manager.io",
				Version: "v1",
				Kind:    "Issuer",
			},
		},
		{
			name:     "custom resource example",
			manifest: []byte(exampleCustomResourceManifest),
			want: &schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "Example",
			},
		},
	}

	scheme := clientgoscheme.Scheme

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			object, got, err := util.RawManifestToObject(scheme, tt.manifest)
			if err != nil {
				t.Fatalf("failed to convert: %s", err)
			}

			if object == nil {
				t.Error("returned object is nil")
			}

			if got == nil {
				t.Error("returned gvk is nil")
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("gvk mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMustRawManifestToObject(t *testing.T) {
	t.Parallel()

	panicked := false

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}

		if !panicked {
			t.Fatal("did not occur panic")
		}
	}()

	util.MustRawManifestToObject(runtime.NewScheme(), []byte("invalid data"))
}

func TestSplitManifests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		manifests []byte
		wants     [][]byte
	}{
		{
			name:      "split yaml",
			manifests: []byte(rawManifests),
			wants: [][]byte{
				[]byte(deploymentManifest),
				[]byte(serviceManifest),
			},
		},
		{
			name: "split json",
			manifests: []byte(
				fmt.Sprintf("[%s,%s]",
					toJSON(t, []byte(deploymentManifest)),
					toJSON(t, []byte(serviceManifest)),
				),
			),
			wants: [][]byte{
				[]byte(deploymentManifest),
				[]byte(serviceManifest),
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gots, err := util.SplitManifests(tt.manifests)
			if err != nil {
				t.Fatalf("failed to split manifests: %s", err)
			}

			set := make(map[string]struct{})

			for _, got := range gots {
				set[string(got)] = struct{}{}
			}

			for _, want := range tt.wants {
				if _, ok := set[string(want)]; ok {
					t.Error("not found expected manifest")
				}
			}
		})
	}
}

func TestSplitYAMLDocument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		manifests []byte
		wants     [][]byte
	}{
		{
			name:      "valid manifests",
			manifests: []byte(rawManifests),
			wants: [][]byte{
				[]byte(deploymentManifest),
				[]byte(serviceManifest),
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			manifests, err := util.SplitYAMLDocument(tt.manifests)
			if err != nil {
				t.Fatalf("failed to split yaml document: %s", err)
			}

			set := make(map[string]struct{})

			for _, got := range manifests {
				set[string(got)] = struct{}{}
			}

			for _, want := range tt.wants {
				if _, ok := set[string(want)]; !ok {
					t.Error("not found yaml document")
				}
			}
		})
	}
}

func TestObjectKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		object runtime.Object
		want   schema.GroupVersionKind
	}{
		{
			name:   "deployment",
			object: toObject(t, []byte(deploymentManifest)),
			want: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
		},
		{
			name:   "service",
			object: toObject(t, []byte(serviceManifest)),
			want: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service",
			},
		},
		{
			name:   "custom resource issuer",
			object: toObject(t, []byte(issuerManifest)),
			want: schema.GroupVersionKind{
				Group:   "cert-manager.io",
				Version: "v1",
				Kind:    "Issuer",
			},
		},
		{
			name:   "custom resource example",
			object: toObject(t, []byte(exampleCustomResourceManifest)),
			want: schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "Example",
			},
		},
	}

	scheme := clientgoscheme.Scheme

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := util.ObjectKinds(scheme, tt.object)
			if err != nil {
				t.Fatalf("failed to convert: %s", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("gvk mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSetGVKIfDoesNotExist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		object  runtime.Object
		want    schema.GroupVersionKind
		wantSet bool
	}{
		{
			name: "hpa v1",
			object: &autoscalingv1.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: schema.GroupVersionKind{
				Group:   "autoscaling",
				Version: "v1",
				Kind:    "HorizontalPodAutoscaler",
			},
			wantSet: true,
		},
		{
			name: "hpa v2beta2",
			object: &autoscalingv2beta2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			want: schema.GroupVersionKind{
				Group:   "autoscaling",
				Version: "v2beta2",
				Kind:    "HorizontalPodAutoscaler",
			},
			wantSet: true,
		},
		{
			name:   "do not set the GVK",
			object: toObject(t, []byte(deploymentManifest)),
			want: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			wantSet: false,
		},
	}

	scheme := clientgoscheme.Scheme

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotSet, err := util.SetGVKIfDoesNotExist(scheme, tt.object)
			if err != nil {
				t.Errorf("failed to set GVK: %s", err)
			}

			if gotSet != tt.wantSet {
				t.Errorf("result of set GVK: got %t, want %t", gotSet, tt.wantSet)
			}

			got := tt.object.GetObjectKind().GroupVersionKind()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("gvk mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func toObject(t *testing.T, data []byte) runtime.Object {
	t.Helper()

	codecFactory := serializer.NewCodecFactory(clientgoscheme.Scheme)
	deserializer := codecFactory.UniversalDeserializer()

	object, _, err := deserializer.Decode(data, nil, nil)
	// If Scheme is not registered, try to convert to unstructured.
	if runtime.IsNotRegisteredError(err) {
		object, _, err = deserializer.Decode(data, nil, &unstructured.Unstructured{})
		if err != nil {
			t.Fatalf("failed to decode manifest to *unstructured.Unstructured: %s", err)
		}
	} else if err != nil {
		t.Fatalf("failed to decode manifest: %s", err)
	}

	return object
}

func toJSON(t *testing.T, data []byte) []byte {
	t.Helper()

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		t.Fatalf("failed to convert to JSON: %s", err)
	}

	return jsonData
}
