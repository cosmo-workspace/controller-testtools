package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"
)

type Header struct {
	SnapshotVersion string `header:"snapshot_version"`
}

func (h *Header) ToString() string {
	return fmt.Sprintf("# chartsnap: snapshot_version=%s\n---\n", h.SnapshotVersion)
}

func ParseHeader(line string) *Header {
	h := Header{}
	ht := reflect.TypeOf(h)
	hv := reflect.ValueOf(&h).Elem()

	split := strings.Split(string([]byte(line)[1:]), " ")
	for _, v := range split {
		s := strings.Split(v, "=")
		if len(s) != 2 {
			continue
		}

		headerName := strings.TrimSpace(s[0])
		headerValue := strings.TrimSpace(s[1])

		for i := 0; i < hv.NumField(); i++ {
			field := ht.Field(i)
			if tag, ok := field.Tag.Lookup("header"); ok && tag == headerName {
				hv.Field(i).SetString(headerValue)
			}
		}
	}
	return &h
}
