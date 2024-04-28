package jsonpatch

import (
	"reflect"
	"testing"
)

func TestDecodePatchKey(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				k: "/metadata/annotations/field~1name",
			},
			want: "/metadata/annotations/field/name",
		},
		{
			name: "test2",
			args: args{
				k: "/metadata/annotations/field~0name",
			},
			want: "/metadata/annotations/field~name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodePatchKey(tt.args.k); got != tt.want {
				t.Errorf("DecodePatchKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitPathDecoded(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				path: "/metadata/annotations/field~1name",
			},
			want: []string{"metadata", "annotations", "field/name"},
		},
		{
			name: "test2",
			args: args{
				path: "/metadata/annotations/field~0name",
			},
			want: []string{"metadata", "annotations", "field~name"},
		},
		{
			name: "test3",
			args: args{
				path: "/metadata",
			},
			want: []string{"metadata"},
		},
		{
			name: "test4",
			args: args{
				path: "/",
			},
			want: []string{""},
		},
		{
			name: "test5",
			args: args{
				path: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitPathDecoded(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitPathDecoded() = %v, want %v", got, tt.want)
			}
		})
	}
}
