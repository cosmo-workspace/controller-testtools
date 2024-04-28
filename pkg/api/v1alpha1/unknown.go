package v1alpha1

import (
	"bytes"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "helm-chartsnap.jlandowner.dev", Version: "v1alpha1"}
)

func NewUnknownError(raw string) *UnknownError {
	return &UnknownError{Raw: raw}
}

type UnknownError struct {
	Raw string
}

func (e *UnknownError) Error() string {
	return fmt.Sprintf("WARN: failed to recognize a resource in stdout/stderr of helm template command output. snapshot it as Unknown: \n---\n%s\n---", e.Raw)
}

func (e *UnknownError) Unstructured() *metaV1.Unstructured {
	return &metaV1.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": GroupVersion.String(),
			"kind":       "Unknown",
			"raw":        e.Raw,
		},
	}
}

func (e *UnknownError) Node() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "apiVersion"},
			{Kind: yaml.ScalarNode, Value: GroupVersion.String()},
			{Kind: yaml.ScalarNode, Value: "kind"},
			{Kind: yaml.ScalarNode, Value: "Unknown"},
			{Kind: yaml.ScalarNode, Value: "raw"},
			{Kind: yaml.ScalarNode, Value: e.Raw, Style: yaml.LiteralStyle},
		},
	}
}

func (e *UnknownError) String() (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err := enc.Encode(e.Node())
	return buf.String(), err
}

func (e *UnknownError) MustString() string {
	s, err := e.String()
	if err != nil {
		panic(fmt.Sprintf("failed to encode UnknownError to YAML: %v", err))
	}
	return s
}
