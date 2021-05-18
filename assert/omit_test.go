package assert_test

/*
func TestOmitMetadata(t *testing.T) {
	tests := []struct {
		name       string
		options    []assert.Option
		object     runtime.Object
		wantObject runtime.Object
	}{
		{
			name: "omit labels",
			options: []assert.Option{
				assert.WithIgnoreHelmManagedLabels(),
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			opt := assert.ToIgnoreOption(tt.options...)

			assert.OmitMetadata(tt.object, opt)
		})
	}
}
*/
