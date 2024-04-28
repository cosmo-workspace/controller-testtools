package v1alpha1

import (
	"reflect"
	"testing"
)

func TestHeader_ToString(t *testing.T) {
	type fields struct {
		Version         string
		SnapshotVersion string
		Chart           string
		Values          string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test Header ToString",
			fields: fields{
				SnapshotVersion: "v3",
			},
			want: "# chartsnap: snapshot_version=v3\n---\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{
				SnapshotVersion: tt.fields.SnapshotVersion,
			}
			if got := h.ToString(); got != tt.want {
				t.Errorf("Header.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want *Header
	}{
		{
			name: "Test ParseHeader",
			args: args{
				line: "# chartsnap: snapshot_version=v3\n",
			},
			want: &Header{
				SnapshotVersion: "v3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseHeader(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
