package yaml

import (
	"io"
	"log/slog"
	"os"
	"testing"

	. "github.com/jlandowner/helm-chartsnap/pkg/snap/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/jlandowner/helm-chartsnap/pkg/api/v1alpha1"
)

func TestYAML(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YAML Suite")
}

var _ = Describe("Decode & Encode", func() {
	It("should success", func() {
		SetLogger(slog.Default())

		// Decode
		manifests := load("testdata/input.yaml")

		// Encode again
		buf, err := Encode(manifests)
		Expect(err).NotTo(HaveOccurred())
		Ω(buf).Should(MatchSnapShot())
	})

	It("should success with converting invalid YAML format", func() {
		SetLogger(slog.Default())

		// Decode
		manifests := load("testdata/input2.yaml")

		// Encode again
		buf, err := Encode(manifests)
		Expect(err).NotTo(HaveOccurred())
		Ω(buf).Should(MatchSnapShot())
	})

	It("should success with converting ScalerNode", func() {
		SetLogger(slog.Default())

		// Decode
		manifests := load("testdata/input3.yaml")

		// Encode again
		buf, err := Encode(manifests)
		Expect(err).NotTo(HaveOccurred())
		Ω(buf).Should(MatchSnapShot())
	})
})

var _ = Describe("ApplyFixedValueToDynamicFieleds", func() {
	It("should replace specified fields", func() {
		cfg := v1alpha1.SnapshotConfig{
			DynamicFields: []v1alpha1.ManifestPath{
				{
					APIVersion: "v1",
					Kind:       "Service",
					Name:       "chartsnap-app2",
					JSONPath: []string{
						"/spec/ports/0/targetPort",
					},
				},
				{
					APIVersion: "v2",
					Kind:       "Service",
					Name:       "chartsnap-app1",
					JSONPath: []string{
						"/spec/ports/0/targetPort",
					},
				},
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
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "chartsnap-app1",
					JSONPath: []string{
						"/metadata/labels/app.kubernetes.io~1version",
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
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "chartsnap-app1",
					JSONPath: []string{
						"/spec/template/spec/serviceAccountName",
					},
					Base64: true,
				},
			},
		}
		manifests := load("testdata/input.yaml")
		err := ApplyFixedValueToDynamicFieleds(cfg, manifests)
		Expect(err).NotTo(HaveOccurred())

		// Encode
		buf, err := Encode(manifests)
		Expect(err).NotTo(HaveOccurred())
		Ω(buf).Should(MatchSnapShot())
	})
})

func load(filePath string) []*yaml.RNode {
	f, err := os.Open(filePath)
	Expect(err).NotTo(HaveOccurred())
	defer f.Close()

	buf, err := io.ReadAll(f)
	Expect(err).NotTo(HaveOccurred())

	manifests, err := Decode(buf)
	Expect(err).NotTo(HaveOccurred())

	return manifests
}
