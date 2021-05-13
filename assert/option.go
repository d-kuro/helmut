package assert

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// option stores the options for assert.
type option struct {
	cmpOptions []cmp.Option

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
// If you want to ignore individual labels, please use the WithIgnoreLabels option.
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

// WithIgnoreLabels is an option to ignore diffs for the specified labels.
// Labels will be ignored regardless of the value of value if the key matches.
func WithIgnoreLabels(labels []string) Option {
	return func(o *option) {
		if o.ignoreOption == nil {
			o.ignoreOption = &ignoreOption{}
		}

		o.ignoreOption.labels = labels
	}
}

// WithIgnoreAnnotations is an option to ignore diffs for the specified annotations.
// Annotations will be ignored regardless of the value of value if the key matches.
func WithIgnoreAnnotations(annotations []string) Option {
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

// omitMetadata will omit metadata based on options to ignore diffs.
// Since the object will be edited, it is recommended to use the deep copied object.
func omitMetadata(object runtime.Object, option *ignoreOption) runtime.Object {
	accessor, err := meta.Accessor(object)
	if err != nil {
		return object
	}

	if option.allHelmManagedLabels {
		omitAllHelmManagedLabels(accessor)
	}

	if len(option.labels) != 0 {
		omitLabels(accessor, option.labels)
	}

	if len(option.annotations) != 0 {
		omitAnnotations(accessor, option.annotations)
	}

	return object
}

// omitLabels will omit the specified labels.
func omitLabels(object metav1.Object, keys []string) {
	labels := object.GetLabels()

	for _, key := range keys {
		delete(labels, key)
	}

	if len(labels) == 0 {
		labels = nil
	}

	object.SetLabels(labels)
}

// omitAnnotations will omit the specified annotation.
func omitAnnotations(object metav1.Object, keys []string) {
	annotations := object.GetAnnotations()

	for _, key := range keys {
		delete(annotations, key)
	}

	if len(annotations) == 0 {
		annotations = nil
	}

	object.SetAnnotations(annotations)
}

// omitAllHelmManagedLabels will omit the labels used by helm.
func omitAllHelmManagedLabels(object metav1.Object) {
	labels := object.GetLabels()

	helmLabels := [...]helmManagedLabel{
		labelAppName,
		labelAppManagedBy,
		labelAppInstance,
		labelAppVersion,
		labelAppComponent,
		labelAppPartOf,
		labelHelmChart,
	}

	for _, helmLabel := range helmLabels {
		delete(labels, helmLabel.String())
	}

	if len(labels) == 0 {
		labels = nil
	}

	object.SetLabels(labels)
}
