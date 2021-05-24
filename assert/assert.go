// Package assert provides a function to compare the object used in the test.
// Use google/go-cmp to compare objects, and output diffs when there are differences.
package assert

import (
	"fmt"

	"github.com/d-kuro/helmut"
	"github.com/d-kuro/helmut/util"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// Contains asserts that the specified manifests contains the specified object.
// If there is a difference in object, fail the test and output diffs.
func Contains(t TestingT, manifests *helmut.Manifests, contains runtime.Object, options ...Option) bool {
	t.Helper()

	opts := &option{}

	for _, o := range options {
		o(opts)
	}

	scheme := manifests.GetScheme()

	if _, err := util.SetGVKIfDoesNotExist(scheme, contains); err != nil {
		t.Errorf("failed to set group,version,kind: %s", err)

		return false
	}

	key, err := helmut.NewObjectKeyFromObject(contains, helmut.WithScheme(scheme))
	if err != nil {
		t.Errorf("failed to create object key: %s", err)

		return false
	}

	actual, key, err := searchObject(manifests, key, opts)
	if err != nil {
		t.Errorf("object was not found")

		return false
	}

	contains = overrideMeta(contains.DeepCopyObject(), key)
	actual = overrideMeta(actual.DeepCopyObject(), key)

	if opts.ignoreOption != nil {
		contains = omitMetadata(contains.DeepCopyObject(), opts.ignoreOption)
		actual = omitMetadata(actual.DeepCopyObject(), opts.ignoreOption)
	}

	for _, fn := range opts.transformers {
		contains = fn(contains.DeepCopyObject())
		actual = fn(actual.DeepCopyObject())
	}

	if diff := cmp.Diff(contains, actual, opts.cmpOptions...); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", key, diff)

		return false
	}

	return true
}

// ContainsWithRawManifest asserts that the specified manifests contains the specified manifest raw data.
// If there is a difference in object, fail the test and output diffs.
func ContainsWithRawManifest(t TestingT, manifests *helmut.Manifests, contains []byte, options ...Option) bool {
	t.Helper()

	scheme := manifests.GetScheme()

	object, _, err := util.RawManifestToObject(scheme, contains)
	if err != nil {
		t.Errorf("failed to convert object: %s", err)

		return false
	}

	return Contains(t, manifests, object, options...)
}

// searchObject searches for objects in manifests.
// If it finds an object, it returns the found object and the key.
func searchObject(
	manifests *helmut.Manifests,
	key helmut.ObjectKey,
	opts *option,
) (runtime.Object, helmut.ObjectKey, error) {
	if object, ok := manifests.Load(key); ok {
		return object, key, nil
	}

	searched := []string{key.String()}

	for _, fn := range opts.additionalKeys {
		addnl := fn(key)

		if object, ok := manifests.Load(addnl); ok {
			return object, addnl, nil
		}

		searched = append(searched, addnl.String())
	}

	return nil, helmut.ObjectKey{}, fmt.Errorf("object %s was not found", searched)
}

// overrideMeta overrides object metadata based on key.
// This function is executed so that no diffs is generated even when using additionalKeys.
func overrideMeta(object runtime.Object, key helmut.ObjectKey) runtime.Object {
	typeAccessor, err := meta.TypeAccessor(object)
	if err != nil {
		return object
	}

	apiVersion, kind := key.GetGroupVersionKind().ToAPIVersionAndKind()

	if typeAccessor.GetKind() != kind {
		typeAccessor.SetKind(kind)
	}

	if typeAccessor.GetAPIVersion() != apiVersion {
		typeAccessor.SetAPIVersion(apiVersion)
	}

	accessor, err := meta.Accessor(object)
	if err != nil {
		return object
	}

	if accessor.GetNamespace() != key.GetNamespace() {
		accessor.SetNamespace(key.GetNamespace())
	}

	if accessor.GetName() != key.GetName() {
		accessor.SetName(key.GetName())
	}

	return object
}
