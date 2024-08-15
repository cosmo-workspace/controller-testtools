package yaml

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/jlandowner/helm-chartsnap/pkg/api/v1alpha1"
	"github.com/jlandowner/helm-chartsnap/pkg/jsonpatch"
)

var (
	logger *slog.Logger
	mutex  sync.Mutex
)

func SetLogger(slogr *slog.Logger) {
	mutex.Lock()
	defer mutex.Unlock()
	logger = slogr
}

func log() *slog.Logger {
	mutex.Lock()
	defer mutex.Unlock()
	if logger == nil {
		logger = slog.Default()
	}
	return logger
}

func Encode(resources []*yaml.RNode) ([]byte, error) {
	// ref: kio.StringAll()
	var b bytes.Buffer
	err := (&kio.ByteWriter{Writer: &b}).Write(resources)
	return b.Bytes(), err
}

func Decode(bs []byte) (out []*yaml.RNode, err error) {
	out, err = decode(bs)
	if err != nil {
		log().Debug(fmt.Sprintf("failed to decode YAML: %v", err))
		out, err = decode(convertInvalidYAMLToUnknown(bs))
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
	}

	err = convertToUnknownNode(out)
	if err != nil {
		return nil, fmt.Errorf("failed to convert scaler node to unknown node: %w", err)
	}
	return out, nil
}

func decode(bs []byte) ([]*yaml.RNode, error) {
	// ref: kio.FromBytes()
	return (&kio.ByteReader{
		OmitReaderAnnotations: true,
		AnchorsAweigh:         true,
		Reader:                bytes.NewBuffer(bs),
	}).Read()
}

func ApplyFixedValueToDynamicFieleds(t v1alpha1.SnapshotConfig, docs []*yaml.RNode) error {
	for _, v := range t.DynamicFields {
		for i, doc := range docs {
			if v.APIVersion == doc.GetApiVersion() &&
				v.Kind == doc.GetKind() &&
				v.Name == doc.GetName() {
				for _, p := range v.JSONPath {
					err := Replace(docs[i], p, v.DynamicValue())
					if err != nil {
						return fmt.Errorf("failed to replace json path: %w", err)
					}
				}
			}
		}
	}
	return nil
}

func convertInvalidYAMLToUnknown(bs []byte) []byte {
	splitString := regexp.MustCompile(`(?m)^---$`).Split(string(bs), -1)

	docs := make([]string, 0, len(splitString))
	for _, v := range splitString {
		if err := yaml.NewDecoder(bytes.NewBufferString(v)).Decode(&yaml.Node{}); err == nil {
			docs = append(docs, v)
		} else {
			unknown := v1alpha1.NewUnknownError(v)
			log().Warn(unknown.Error())
			docs = append(docs, unknown.MustString())
		}
	}
	return []byte(strings.Join(docs, "\n---\n"))
}

func convertToUnknownNode(docs []*yaml.RNode) error {
	for i, v := range docs {
		if v.IsStringValue() {
			vv := v.YNode().Value
			unknown := v1alpha1.NewUnknownError(vv)
			log().Warn(unknown.Error())
			docs[i] = yaml.NewRNode(unknown.Node())
			docs[i].ShouldKeep = true
		} else if v.GetApiVersion() == "" && v.GetKind() == "" {
			vv, _ := Encode([]*yaml.RNode{v})
			unknown := v1alpha1.NewUnknownError(string(vv))
			log().Warn(unknown.Error())
			docs[i] = yaml.NewRNode(unknown.Node())
			docs[i].ShouldKeep = true
		}
	}
	return nil
}

// Replace replaces the value at the given jsonpath with the given value.
// jsonpath is a RFC 6901 JSON path.
func Replace(doc *yaml.RNode, path, value string) error {
	return doc.PipeE(yaml.LookupCreate(yaml.ScalarNode, (jsonpatch.SplitPathDecoded(path))...), yaml.Set(yaml.NewStringRNode(value)))
}
