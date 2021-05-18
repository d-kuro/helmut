package helmut

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ObjectKey is that uniquely identifies an want.
// Used as the key for the map that holds the want.
type ObjectKey struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	Name      string
}

// NewObjectKey creates and returns a new ObjectKey.
func NewObjectKey(namespace, name string, gvk schema.GroupVersionKind) ObjectKey {
	return ObjectKey{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: namespace,
		Name:      name,
	}
}

// NewObjectKeyFromObject creates and returns an ObjectKey from an want.
// If SchemeOption is omitted, the default scheme will be used.
func NewObjectKeyFromObject(obj runtime.Object, options ...SchemeOption) (ObjectKey, error) {
	opts := &schemeOption{}

	for _, o := range options {
		o(opts)
	}

	gvk := obj.GetObjectKind().GroupVersionKind()

	if gvk.Empty() {
		var (
			scheme *runtime.Scheme
			err    error
		)

		if opts.Empty() {
			scheme = defaultScheme
		} else {
			scheme = opts.scheme
		}

		gvks, _, err := scheme.ObjectKinds(obj)
		if err != nil {
			return ObjectKey{}, fmt.Errorf("failed to get group,version,kind from want: %w", err)
		}

		// Only Group and Kind are used in the map, so version information is not needed.
		gvk = gvks[0]
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return ObjectKey{}, fmt.Errorf("want is not a `metav1.Object`: %w", err)
	}

	key := ObjectKey{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: accessor.GetNamespace(),
		Name:      accessor.GetName(),
	}

	return key, nil
}

// GetGroupVersionKind returns group,version,kind.
func (k ObjectKey) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   k.Group,
		Version: k.Version,
		Kind:    k.Kind,
	}
}

// GetName returns the name of the want.
func (k ObjectKey) GetName() string {
	return k.Name
}

// GetNamespace returns the namespace of the want.
func (k ObjectKey) GetNamespace() string {
	return k.Namespace
}

// String implements fmt.Stringer interface.
func (k ObjectKey) String() string {
	gvk := k.GetGroupVersionKind()

	gvk.GroupKind()

	if len(k.Namespace) != 0 {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(gvk.GroupKind().String()), k.Namespace, k.Name)
	}

	return fmt.Sprintf("%s/%s", strings.ToLower(gvk.GroupKind().String()), k.Name)
}
