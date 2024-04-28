package v1alpha1

import (
	"encoding/base64"
	"fmt"
	"os"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

func FromFile[T SnapshotValues | SnapshotConfig](filePath string, out *T) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(out)
	if err != nil {
		return fmt.Errorf("failed to decode file '%s': %w", filePath, err)
	}
	return nil
}

type SnapshotValues struct {
	TestSpec SnapshotConfig `yaml:"testSpec,omitempty"`
}

type SnapshotConfig struct {
	DynamicFields []ManifestPath `yaml:"dynamicFields,omitempty"`
}

type ManifestPath struct {
	Kind       string   `yaml:"kind,omitempty"`
	APIVersion string   `yaml:"apiVersion,omitempty"`
	Name       string   `yaml:"name,omitempty"`
	JSONPath   []string `yaml:"jsonPath,omitempty"`
	Base64     bool     `yaml:"base64,omitempty"`
}

const DynamicValue = "###DYNAMIC_FIELD###"

var Base64DynamicValue = base64.StdEncoding.EncodeToString([]byte(DynamicValue))

func (v *ManifestPath) DynamicValue() string {
	if v.Base64 {
		return Base64DynamicValue
	} else {
		return DynamicValue
	}
}

// Merge merges the snapshot configs into the current snapshot config
// The current snapshot config has higher priority than the given snapshot config
func (t *SnapshotConfig) Merge(cfg SnapshotConfig) {
	// For DynamicFields, it doesn't matter if the same field is replaced with a fixed value several times
	// But the current snapshot config has higher priority than the given snapshot config
	t.DynamicFields = append(cfg.DynamicFields, t.DynamicFields...)
}
