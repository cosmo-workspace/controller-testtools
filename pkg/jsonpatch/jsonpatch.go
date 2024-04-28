package jsonpatch

import "strings"

// From http://tools.ietf.org/html/rfc6901#section-4 :
//
// Evaluation of each reference token begins by decoding any escaped
// character sequence.  This is performed by first transforming any
// occurrence of the sequence '~1' to '/', and then transforming any
// occurrence of the sequence '~0' to '~'.

var (
	RFC6901Decoder = strings.NewReplacer("~1", "/", "~0", "~")
)

func DecodePatchKey(k string) string {
	return RFC6901Decoder.Replace(k)
}

func SplitPathDecoded(path string) []string {
	split := strings.Split(path, "/")
	if len(split) < 2 {
		return nil
	}
	parts := split[1:]
	for i := 0; i < len(parts); i++ {
		parts[i] = DecodePatchKey(parts[i])
	}
	return parts
}
