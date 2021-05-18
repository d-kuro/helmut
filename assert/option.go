package assert

import (
	"github.com/google/go-cmp/cmp"
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
