// Package assert provides a function to compare the object used in the test.
// Use google/go-cmp to compare objects, and output diffs when there are differences.
package assert

import (
	"github.com/d-kuro/helmut"
	"github.com/d-kuro/helmut/util"
	"github.com/google/go-cmp/cmp"
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

	actual, ok := manifests.Load(key)
	if !ok {
		t.Errorf("%s was not found", key)

		return false
	}

	if opts.ignoreOption != nil {
		contains = omitMetadata(contains.DeepCopyObject(), opts.ignoreOption)
		actual = omitMetadata(actual.DeepCopyObject(), opts.ignoreOption)
	}

	if opts.transformOption != nil {
		for _, fn := range opts.transformOption.transformers {
			contains = fn(contains.DeepCopyObject())
			actual = fn(actual.DeepCopyObject())
		}
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
