package util_test

import (
	"testing"

	"github.com/d-kuro/helmut/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
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

func TestRawManifestToUnstructuredObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		manifest  []byte
		wantGVK   schema.GroupVersionKind
		wantError bool
	}{
		{
			name:     "custom resource Issuer",
			manifest: []byte(issuerManifest),
			wantGVK: schema.GroupVersionKind{
				Group:   "cert-manager.io",
				Version: "v1",
				Kind:    "Issuer",
			},
			wantError: false,
		},
		{
			name:     "custom resource Example",
			manifest: []byte(exampleCustomResourceManifest),
			wantGVK: schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "Example",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			object, gvk, err := util.RawManifestToUnstructuredObject(tt.manifest)
			if err != nil {
				t.Fatalf("failed to convert: %s", err)
			}

			if object == nil {
				t.Error("returned object is nil")
			}

			if gvk == nil {
				t.Error("returned gvk is nil")
			}
		})
	}
}

func TestMustRawManifestToUnstructuredObject(t *testing.T) {
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

	util.MustRawManifestToUnstructuredObject([]byte("invalid data"))
}
