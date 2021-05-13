// Package util provides utilities for testing.
package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// ObjectKinds returns group,version,kind of the object.
// It may be possible to get multiple group,version,kind in which case error will be returned.
func ObjectKinds(scheme *runtime.Scheme, object runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("failed to get group,version,kind from object: %w", err)
	}

	if len(gvks) > 1 {
		return schema.GroupVersionKind{}, fmt.Errorf("multiple group,version,kind returned: %#v", gvks)
	}

	gvk := gvks[0]

	return gvk, nil
}

// SetGVKIfDoesNotExist sets group,version,kind if they do not exist in object.
// If set, returns true.
func SetGVKIfDoesNotExist(scheme *runtime.Scheme, object runtime.Object) (bool, error) {
	accessor, err := meta.TypeAccessor(object)
	if err != nil {
		return false, fmt.Errorf("could not get accessor to TypeMeta: %w", err)
	}

	if len(accessor.GetAPIVersion()) != 0 && len(accessor.GetKind()) != 0 {
		return false, nil
	}

	gvk, err := ObjectKinds(scheme, object)
	if err != nil {
		return false, fmt.Errorf("failed to get group,version,kind from object: %w", err)
	}

	apiVersion, kind := gvk.ToAPIVersionAndKind()

	if len(accessor.GetAPIVersion()) == 0 {
		accessor.SetAPIVersion(apiVersion)
	}

	if len(accessor.GetKind()) == 0 {
		accessor.SetKind(kind)
	}

	return true, nil
}

// RawManifestToObject converts a raw manifest to a object.
func RawManifestToObject(scheme *runtime.Scheme, data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	codecFactory := serializer.NewCodecFactory(scheme)
	deserializer := codecFactory.UniversalDeserializer()

	object, gvk, err := deserializer.Decode(data, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return object, gvk, nil
}

// MustRawManifestToObject calls RawManifestToObject and panics if an error exists.
func MustRawManifestToObject(scheme *runtime.Scheme, data []byte) (runtime.Object, *schema.GroupVersionKind) {
	object, gvk, err := RawManifestToObject(scheme, data)
	if err != nil {
		panic(err)
	}

	return object, gvk
}

// RawManifestToUnstructuredObject converts a raw manifest to a unstructured object.
func RawManifestToUnstructuredObject(data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	codecFactory := serializer.NewCodecFactory(runtime.NewScheme())
	deserializer := codecFactory.UniversalDeserializer()

	object, gvk, err := deserializer.Decode(data, nil, &unstructured.Unstructured{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode manifest to *unstructured.Unstructured: %w", err)
	}

	return object, gvk, nil
}

// MustRawManifestToUnstructuredObject calls RawManifestToUnstructuredObject and panics if an error exists.
func MustRawManifestToUnstructuredObject(data []byte) (runtime.Object, *schema.GroupVersionKind) {
	object, gvk, err := RawManifestToUnstructuredObject(data)
	if err != nil {
		panic(err)
	}

	return object, gvk
}
