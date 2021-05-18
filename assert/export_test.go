package assert

// OmitMetadata exports the omitMetadata function for testing.
var OmitMetadata = omitMetadata

// IgnoreOption exports the ignoreOption for testing.
type IgnoreOption = ignoreOption

// SetOptions sets the option.
func (o *IgnoreOption) SetOptions(options ...Option) {
	opts := &option{}

	for _, o := range options {
		o(opts)
	}

	o.labels = opts.ignoreOption.labels
	o.annotations = opts.ignoreOption.annotations
	o.allHelmManagedLabels = opts.ignoreOption.allHelmManagedLabels
}
