package assert

import (
	"github.com/d-kuro/helmut"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
)

// option stores the options for assert.
type option struct {
	// cmpOptions is the option used when calling cmp.Diff.
	cmpOptions []cmp.Option

	// transformers is the function used to transform an object.
	transformers []func(runtime.Object) runtime.Object

	// additionalKeys is a function that generates additional search keys to be executed if the object is not found.
	// The previous search key is passed as an args, you can overwrite it.
	additionalKeys []func(helmut.ObjectKey) helmut.ObjectKey

	// ignoreOption stores the option to ignore object diffs.
	ignoreOption *ignoreOption
}

// ignoreOption stores the option to ignore object diffs.
type ignoreOption struct {
	allHelmManagedLabels bool
	labels               []string
	annotations          []string
}

// Option is the option used when asserting.
type Option func(*option)

// WithIgnoreHelmManagedLabels is an option to ignore diffs
// in labels that Helm is supposed to use in general.
// Labels will be ignored regardless of the value of value if the key matches.
// If you want to ignore individual labels, please use the WithIgnoreLabelKeys option.
//
// The labels that are ignored are:
//   app.kubernetes.io/name
//   app.kubernetes.io/managed-by
//   app.kubernetes.io/instance
//   app.kubernetes.io/version
//   app.kubernetes.io/component
//   app.kubernetes.io/part-of
//   helm.sh/chart
//
// see: https://helm.sh/docs/chart_best_practices/labels/
func WithIgnoreHelmManagedLabels() Option {
	return func(o *option) {
		if o.ignoreOption == nil {
			o.ignoreOption = &ignoreOption{}
		}

		o.ignoreOption.allHelmManagedLabels = true
	}
}

// WithIgnoreLabelKeys is an option to ignore diffs for the specified labels.
// Labels will be ignored regardless of the value of value if the key matches.
func WithIgnoreLabelKeys(labels ...string) Option {
	return func(o *option) {
		if o.ignoreOption == nil {
			o.ignoreOption = &ignoreOption{}
		}

		o.ignoreOption.labels = labels
	}
}

// WithIgnoreAnnotationKeys is an option to ignore diffs for the specified annotations.
// Annotations will be ignored regardless of the value of value if the key matches.
func WithIgnoreAnnotationKeys(annotations ...string) Option {
	return func(o *option) {
		if o.ignoreOption == nil {
			o.ignoreOption = &ignoreOption{}
		}

		o.ignoreOption.annotations = annotations
	}
}

// WithCmpOptions specifies the options to be used when comparing objects with google/go-cmp.
func WithCmpOptions(opts ...cmp.Option) Option {
	return func(o *option) {
		o.cmpOptions = append(o.cmpOptions, opts...)
	}
}

// WithTransformer is an option to provide a function to freely transform the object to be compared.
// For example, you can use it to omit or edit a particular field.
// The function passed here will be executed just before the comparison
// and will be applied to both of the two Objects being compared.
//
// Example of omitting the securityContext of a Pod:
//
//  omitSecurityContext := func(obj runtime.Object) runtime.Object {
//  	pod, ok := obj.(*corev1.Pod)
//  	if !ok {
//  		return obj
//  	}
//
//  	pod.Spec.SecurityContext = nil
//
//  	return pod
//  }
//
//  assert.Contains(t, manifests, obj, assert.WithTransformer(omitSecurityContext))
//
func WithTransformer(fn ...func(runtime.Object) runtime.Object) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, fn...)
	}
}

// WithAdditionalKeys you can specify a function to generate additional search keys
// that will be used if the object is not found.
// The original search key is passed as an argument to the function, you can be overwritten.
//
// When an object is found by the generated key,
// the group, version, kind, name, namespace of the object are rewritten based on the key.
// This will prevent the above fields from diffs.
//
// You can pass multiple functions as arguments,
// but when the generated key finds the object,
// the execution of the remaining functions
// will be interrupted and the object comparison process will be performed.
//
// For example, it can be used to ignore the release name given by Helm.
//
// Example of removing prefix from name:
//
//  removePrefix := func(key helmut.ObjectKey) helmut.ObjectKey {
//  	if strings.HasPrefix(key.GetName(), "prefix") {
//  		key.Name = strings.TrimPrefix(key.GetName(), "prefix-")
//  	}
//  	return key
//  }
//
//  assert.Contains(t, manifests, obj, assert.WithAdditionalKeys(removePrefix))
//
func WithAdditionalKeys(fn ...func(helmut.ObjectKey) helmut.ObjectKey) Option {
	return func(o *option) {
		o.additionalKeys = append(o.additionalKeys, fn...)
	}
}

// helmManagedLabel is the label used by Helm.
// see: https://helm.sh/docs/chart_best_practices/labels/
type helmManagedLabel string

const (
	labelAppName      helmManagedLabel = "app.kubernetes.io/name"
	labelAppManagedBy helmManagedLabel = "app.kubernetes.io/managed-by"
	labelAppInstance  helmManagedLabel = "app.kubernetes.io/instance"
	labelAppVersion   helmManagedLabel = "app.kubernetes.io/version"
	labelAppComponent helmManagedLabel = "app.kubernetes.io/component"
	labelAppPartOf    helmManagedLabel = "app.kubernetes.io/part-of"
	labelHelmChart    helmManagedLabel = "helm.sh/chart"
)

// String implements the fmt.Stringer interface.
func (l helmManagedLabel) String() string {
	return string(l)
}
