package helmut

// option stores the template options.
type option struct {
	namespace   string
	apiVersions []string
	includeCRDs bool

	// value options
	valueFiles   []string
	stringValues []string
	values       []string
	fileValues   []string
}

// Option is an option to specify when calling RenderTemplates.
type Option func(*option)

// WithNamespace specifies the namespace.
// This is equivalent to the "--namespace" option of the "helm template" command.
func WithNamespace(namespace string) Option {
	return func(o *option) {
		o.namespace = namespace
	}
}

// WithAPIVersions specifies the apiVersions.
// This is equivalent to the "--api-versions" option of the "helm template" command.
func WithAPIVersions(apiVersions []string) Option {
	return func(o *option) {
		o.apiVersions = apiVersions
	}
}

// WithIncludeCRDs will include CRDs in the templated output.
// This is equivalent to the "--include-crds" option of the "helm template" command.
func WithIncludeCRDs() Option {
	return func(o *option) {
		o.includeCRDs = true
	}
}

// WithValues specifies values in a YAML file or a URL (can specify multiple).
// This is equivalent to the "--values" or "-f" option of the "helm template" command.
func WithValues(files []string) Option {
	return func(o *option) {
		o.valueFiles = files
	}
}

// WithSetString set STRING values just like the command line.
// (can specify multiple or separate values with commas: key1=val1,key2=val2)
// This is equivalent to the "--set-string" option of the "helm template" command.
func WithSetString(values []string) Option {
	return func(o *option) {
		o.stringValues = values
	}
}

// WithSet set values just like the command line.
// (can specify multiple or separate values with commas: key1=val1,key2=val2)
// This is equivalent to the "--set" option of the "helm template" command.
func WithSet(values []string) Option {
	return func(o *option) {
		o.values = values
	}
}

// WithSetFile set values just like the command line.
// (can specify multiple or separate values with commas: key1=val1,key2=val2)
// This is equivalent to the "set-file" option of the "helm template" command.
func WithSetFile(files []string) Option {
	return func(o *option) {
		o.fileValues = files
	}
}
