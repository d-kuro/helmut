package assert_test

import (
	"testing"

	"github.com/d-kuro/helmut/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSortObjectsByNameField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		v         interface{}
		sorted    interface{}
		wantEqual bool
	}{
		{
			name:      "sort name field",
			v:         []struct{ Name string }{{"foo"}, {"bar"}},
			sorted:    []struct{ Name string }{{"bar"}, {"foo"}},
			wantEqual: true,
		},
		{
			name:      "already sorted",
			v:         []struct{ Name string }{{"foo"}, {"bar"}},
			sorted:    []struct{ Name string }{{"foo"}, {"bar"}},
			wantEqual: true,
		},
		{
			name:      "name field missing",
			v:         []struct{ V string }{{"foo"}, {"bar"}},
			sorted:    []struct{ V string }{{"bar"}, {"foo"}},
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotEqual := cmp.Equal(tt.sorted, tt.v, cmpopts.SortSlices(assert.SortObjectsByNameField))
			if gotEqual != tt.wantEqual {
				if diff := cmp.Diff(tt.sorted, tt.v, cmpopts.SortSlices(assert.SortObjectsByNameField)); diff != "" {
					t.Errorf("equal: %v, want: %v, mismatch (-want +got):\n%s", gotEqual, tt.wantEqual, diff)
				} else {
					t.Errorf("equal: %v, want: %v", gotEqual, tt.wantEqual)
				}
			}
		})
	}
}
