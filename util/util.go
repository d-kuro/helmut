// Package util provides utilities for testing.
package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
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
// Attempts to convert to unstructured.Unstructured if no type is registered in scheme.
func RawManifestToObject(scheme *runtime.Scheme, data []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	codecFactory := serializer.NewCodecFactory(scheme)
	deserializer := codecFactory.UniversalDeserializer()

	object, gvk, err := deserializer.Decode(data, nil, nil)
	// If Scheme is not registered, try to convert to unstructured.
	if runtime.IsNotRegisteredError(err) {
		object, gvk, err = deserializer.Decode(data, nil, &unstructured.Unstructured{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode manifest to *unstructured.Unstructured: %w", err)
		}
	} else if err != nil {
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

// SplitManifests takes a single large manifest and splits it into individual manifests.
// Both JSON and YAML are supported, but the returned manifests will be converted to JSON.
func SplitManifests(data []byte) ([][]byte, error) {
	var result [][]byte

	r := bytes.NewReader(data)
	decoder := yaml.NewYAMLOrJSONDecoder(r, 4096)

	for {
		ext := runtime.RawExtension{}
		if err := decoder.Decode(&ext); err != nil {
			if errors.Is(err, io.EOF) {
				return result, nil
			}

			if err != nil {
				return nil, fmt.Errorf("failed to parse manifests: %w", err)
			}
		}

		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			continue
		}

		result = append(result, ext.Raw)
	}
}

// SplitYAMLDocument is a splitting YAML document into individual documents.
func SplitYAMLDocument(data []byte) ([][]byte, error) {
	bufReader := bufio.NewReader(bytes.NewReader(data))
	yamlReader := yaml.NewYAMLReader(bufReader)

	var result [][]byte

	for {
		document, err := yamlReader.Read()
		if errors.Is(err, io.EOF) {
			return result, nil
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read YAML document: %w", err)
		}

		document = bytes.TrimSpace(document)

		result = append(result, document)
	}
}
