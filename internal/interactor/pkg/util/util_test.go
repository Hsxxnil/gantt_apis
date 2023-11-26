package util

import (
	"reflect"
	"testing"
)

func TestRemoveString(t *testing.T) {
	type args struct {
		s    []string
		item string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "remove string",
			args: args{
				s:    []string{"a", "b", "c"},
				item: "b",
			},
			want: []string{"a", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveString(tt.args.s, tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveString() = %v, want %v", got, tt.want)
			}
		})
	}
}
