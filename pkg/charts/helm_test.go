package charts

import (
	"context"

	. "github.com/jlandowner/helm-chartsnap/pkg/snap/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helm", func() {
	Context("when Execute", func() {
		It("should execute with expected args and env", func() {
			o := &HelmTemplateCmdOptions{
				HelmPath:    "./testdata/helm_cmd.bash",
				ReleaseName: "aaa",
				Namespace:   "bbb",
				Chart:       "ccc",
				ValuesFile:  "ddd",
			}

			out, err := o.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(MatchSnapShot())
		})
	})

	Context("when Execute without namespace", func() {
		It("should execute with expected args and env", func() {
			o := &HelmTemplateCmdOptions{
				HelmPath:    "./testdata/helm_cmd.bash",
				ReleaseName: "chartsnap",
				Chart:       "charts/app1/",
				ValuesFile:  "charts/app1/test/test.values.yaml",
			}

			out, err := o.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(MatchSnapShot())
		})
	})

	Context("when Execute without values", func() {
		It("should execute with expected args and env", func() {
			o := &HelmTemplateCmdOptions{
				HelmPath:    "./testdata/helm_cmd.bash",
				ReleaseName: "chartsnap",
				Namespace:   "default",
				Chart:       "charts/app1/",
			}

			out, err := o.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(MatchSnapShot())
		})
	})

	Context("when Execute with additional args", func() {
		It("should execute with expected args and env", func() {
			o := &HelmTemplateCmdOptions{
				HelmPath:       "./testdata/helm_cmd.bash",
				ReleaseName:    "chartsnap",
				Namespace:      "xxx",
				Chart:          "postgres",
				ValuesFile:     "postgres.values.yaml",
				AdditionalArgs: []string{"--repo", "https://charts.bitnami.com/bitnami", "--skip-tests"},
			}

			out, err := o.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(MatchSnapShot())
		})
	})

	Context("test mocks", func() {
		It("should execute as helm cmd", func() {
			o := &HelmTemplateCmdOptions{
				HelmPath:    "./testdata/helm_empty.bash",
				ReleaseName: "release",
				Chart:       "chart",
			}
			out, err := o.Execute(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(MatchSnapShot())
		})
	})
})
