package unstructured

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlUtil "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yaml "sigs.k8s.io/yaml/goyaml.v3"

	"github.com/jlandowner/helm-chartsnap/pkg/api/v1alpha1"
)

var logger *slog.Logger

func SetLogger(slogr *slog.Logger) {
	logger = slogr
}

func log() *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return logger
}

func Encode(arr []metaV1.Unstructured) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	for _, d := range arr {
		if err := enc.Encode(d.Object); err != nil {
			return nil, fmt.Errorf("failed to encode unstructured to YAML: %w", err)
		}
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	return buf.Bytes(), nil
}

func Decode(source string) ([]metaV1.Unstructured, []error) {
	splitString := regexp.MustCompile(`(?m)^---$`).Split(source, -1)
	resources := make([]metaV1.Unstructured, 0, len(splitString))

	var errs []error = make([]error, 0)
	for _, v := range splitString {
		if strings.TrimSpace(v) == "" {
			continue
		}
		// split lines and if all lines start with '#', it is skipped
		empty := true
		for _, line := range strings.Split(v, "\n") {
			l := strings.TrimPrefix(line, " ")
			l = strings.TrimPrefix(l, "\t")
			if l != "" && !strings.HasPrefix(l, "#") {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		_, obj, err := StringToUnstructured(v)
		if err != nil {
			err := v1alpha1.NewUnknownError(v)
			obj = err.Unstructured()
			errs = append(errs, err)
		}
		resources = append(resources, *obj)
	}

	return resources, errs
}

func ApplyFixedValue(t v1alpha1.SnapshotConfig, manifests []metaV1.Unstructured) error {
	for _, v := range t.DynamicFields {
		for i, obj := range manifests {
			if v.APIVersion == obj.GetAPIVersion() &&
				v.Kind == obj.GetKind() &&
				v.Name == obj.GetName() {
				for _, p := range v.JSONPath {
					newObj, err := Replace(manifests[i], p, v.DynamicValue())
					if err != nil {
						return fmt.Errorf("failed to replace json path: %w", err)
					}
					manifests[i] = *newObj
				}
			}
		}
	}
	return nil
}

func Replace(obj metaV1.Unstructured, key, value string) (*metaV1.Unstructured, error) {
	str_patch := fmt.Sprintf(`[{"op": "replace", "path": "%s", "value": "%s"}]`, key, value)
	bytes_obj, err := UnstructuredToJSONBytes(obj.DeepCopy())
	if err != nil {
		return nil, fmt.Errorf("failed to encode unstructured to JSON: %w", err)
	}

	// patch JSON6902
	patch, err := jsonpatch.DecodePatch([]byte(str_patch))
	if err != nil {
		return nil, fmt.Errorf("failed to decode patch: %w: %s", err, str_patch)
	}

	patched, err := patch.Apply(bytes_obj)
	if err != nil {
		return nil, fmt.Errorf("failed to patch JSON6902: %w: patch=%s: buf=%s", err, str_patch, bytes_obj)
	}

	_, patchedObj, err := BytesToUnstructured(patched)

	return patchedObj, err
}

func StringToUnstructured(data string) (*schema.GroupVersionKind, *metaV1.Unstructured, error) {
	return BytesToUnstructured([]byte(data))
}

func BytesToUnstructured(data []byte) (*schema.GroupVersionKind, *metaV1.Unstructured, error) {
	obj := &metaV1.Unstructured{}
	dec := yamlUtil.NewDecodingSerializer(metaV1.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(data), nil, obj)
	if err != nil {
		return nil, nil, err
	}
	return gvk, obj, nil
}

func UnstructuredToJSONBytes(obj *metaV1.Unstructured) ([]byte, error) {
	return json.Marshal(obj)
}
