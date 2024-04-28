package v1alpha1

import (
	"testing"

	. "github.com/jlandowner/helm-chartsnap/pkg/snap/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCharts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Charts Suite")
}

var _ = Describe("TestSpec", func() {
	var _ = Describe("FromFile", func() {
		Context("when values.yaml has testSpec", func() {
			It("should load config", func() {
				var v SnapshotValues
				err := FromFile("testdata/testspec_values.yaml", &v)
				Expect(err).NotTo(HaveOccurred())
				Expect(v).To(MatchSnapShot())
			})
		})

		Context("when loading .chartsnap.yaml", func() {
			It("should load config", func() {
				var v SnapshotConfig
				err := FromFile("testdata/.chartsnap.yaml", &v)
				Expect(err).NotTo(HaveOccurred())
				Expect(v).To(MatchSnapShot())
			})
		})

		Context("when loading invalid yaml", func() {
			It("should not load config", func() {
				var v SnapshotValues
				err := FromFile("testdata/testspec_values_invalid.yaml", &v)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).To(MatchSnapShot())
			})
		})
	})

	var _ = Describe("Merge", func() {
		It("should merge dynamic fields", func() {
			cfg1 := SnapshotConfig{
				DynamicFields: []ManifestPath{
					{
						APIVersion: "v1",
						Kind:       "service",
						Name:       "chartsnap-app1",
						JSONPath: []string{
							"/spec/ports/0/targetPort",
						},
					},
					{
						APIVersion: "v1",
						Kind:       "Service",
						Name:       "chartsnap-app1",
						JSONPath: []string{
							"/spec/ports/1/targetPort",
						},
					},
				},
			}
			cfg2 := SnapshotConfig{
				DynamicFields: []ManifestPath{
					{
						APIVersion: "v1",
						Kind:       "service",
						Name:       "chartsnap-app1",
						JSONPath: []string{
							"/spec/ports/0/targetPort",
						},
					},
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "chartsnap-app1-test-connection",
						JSONPath: []string{
							"/metadata/name",
						},
					},
				},
			}
			cfg1.Merge(cfg2)
			Expect(cfg1).To(MatchSnapShot())
		})
	})

})

func TestManifestPath_DynamicValue(t *testing.T) {
	type fields struct {
		Kind       string
		APIVersion string
		Name       string
		JSONPath   []string
		Base64     bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test DynamicValue",
			fields: fields{
				Base64: false,
			},
			want: "###DYNAMIC_FIELD###",
		},
		{
			name: "Test DynamicValue Base64",
			fields: fields{
				Base64: true,
			},
			want: "IyMjRFlOQU1JQ19GSUVMRCMjIw==",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &ManifestPath{
				Kind:       tt.fields.Kind,
				APIVersion: tt.fields.APIVersion,
				Name:       tt.fields.Name,
				JSONPath:   tt.fields.JSONPath,
				Base64:     tt.fields.Base64,
			}
			if got := v.DynamicValue(); got != tt.want {
				t.Errorf("ManifestPath.DynamicValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
