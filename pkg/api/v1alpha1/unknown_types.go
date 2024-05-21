// +kubebuilder:object:generate=true
// +groupName=helm-chartsnap.jlandowner.dev
package v1alpha1

import (
	"bytes"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "helm-chartsnap.jlandowner.dev", Version: "v1alpha1"}
)

func NewUnknownError(raw string) *Unknown {
	return &Unknown{Raw: raw}
}

// +kubebuilder:object:root=true
// Unknown is a placeholder for an unrecognized resource in stdout/stderr of helm template command output.
type Unknown struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Raw is the raw string of the helm output.
	Raw string `json:"raw,omitempty"`
}

func (e *Unknown) Error() string {
	return fmt.Sprintf("failed to recognize a resource in stdout/stderr of helm template command output. snapshot it as Unknown: \n---\n%s\n---", e.Raw)
}

func (e *Unknown) Unstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": GroupVersion.String(),
			"kind":       "Unknown",
			"raw":        e.Raw,
		},
	}
}

func (e *Unknown) Node() *yaml.Node {
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

func (e *Unknown) String() (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err := enc.Encode(e.Node())
	return buf.String(), err
}

func (e *Unknown) MustString() string {
	s, err := e.String()
	if err != nil {
		panic(fmt.Sprintf("failed to encode Unknown to YAML: %v", err))
	}
	return s
}
