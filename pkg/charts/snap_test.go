package charts

import (
	"context"
	"os"
	"testing"

	. "github.com/jlandowner/helm-chartsnap/pkg/snap/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jlandowner/helm-chartsnap/pkg/api/v1alpha1"
)

func TestCharts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Charts Suite")
}

var _ = Describe("Snap", func() {
	Context("v3 snapshot matched", func() {
		It("should return success response", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v3.yaml",
				SnapshotVersion:  "v3",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
			Expect(res.FailureMessage).To(MatchSnapShot())
		})
	})

	Context("v2 snapshot matched", func() {
		It("should return success response", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v2.yaml",
				SnapshotVersion:  "v2",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
			Expect(res.FailureMessage).To(MatchSnapShot())
		})
	})

	Context("v1 snapshot matched", func() {
		It("should return success response", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v1.toml",
				SnapshotVersion:  "v1",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
			Expect(res.FailureMessage).To(MatchSnapShot())
		})
	})

	Context("v2 snapshot not matched", func() {
		It("should return unmatched response", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_unmatch_v2.yaml",
				SnapshotVersion:  "v2",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeFalseBecause("diff: %s", res.FailureMessage))
			Expect(res.FailureMessage).To(MatchSnapShot())
		})
	})

	Context("v3 snapshot not matched", func() {
		BeforeEach(func() {
			copyFile := func(src, dst string) {
				srcFile, err := os.Open(src)
				Expect(err).NotTo(HaveOccurred())
				defer srcFile.Close()

				dstFile, err := os.Create(dst)
				Expect(err).NotTo(HaveOccurred())
				defer dstFile.Close()

				_, err = srcFile.WriteTo(dstFile)
				Expect(err).NotTo(HaveOccurred())
			}
			copyFile("__snapshots__/helm_stub_snap_unmatch_v3.yaml", "__snapshots__/helm_stub_snap_unmatch_v3_copy.yaml")
		})
		AfterEach(func() {
			os.Remove("__snapshots__/helm_stub_snap_unmatch_v3_copy.yaml")
		})

		It("should return unmatched response", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_unmatch_v3_copy.yaml",
				SnapshotVersion:  "v3",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeFalseBecause("diff: %s", res.FailureMessage))
			Expect(res.FailureMessage).To(MatchSnapShot())
		})

		It("should update snapshot", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_unmatch_v3_copy.yaml",
				SnapshotVersion:  "v3",
				DiffContextLineN: 3,
				UpdateSnapshot:   true,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})

	Context("empty snapshot", func() {
		It("should be successfull and no error occers (after v3)", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath: "./testdata/helm_empty.bash",
				},
				SnapshotFile:     "__snapshots__/empty.yaml",
				SnapshotVersion:  "",
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})

	Context("snapshot version is not specified", func() {
		It("should match with latest snapshot format version", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_latest.yaml",
				SnapshotVersion:  "", // not specified
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})

	Context("snapshot version is not specified and snapshot file format is v1", func() {
		It("should match with v1 snapshot format", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v1.toml",
				SnapshotVersion:  "", // not specified
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})

	Context("snapshot version is not specified and snapshot file format is v2", func() {
		It("should match with v1 snapshot format", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v2.yaml",
				SnapshotVersion:  "", // not specified
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})

	Context("snapshot version is not specified and snapshot file format is v3", func() {
		It("should match with v1 snapshot format", func() {
			ss := &ChartSnapshotter{
				HelmTemplateCmdOptions: HelmTemplateCmdOptions{
					HelmPath:    "./testdata/helm_stub.bash",
					ReleaseName: "aaa",
					Namespace:   "bbb",
					Chart:       "ccc",
					ValuesFile:  "./testdata/snap_values.yaml",
				},
				SnapshotConfig: v1alpha1.SnapshotConfig{
					DynamicFields: []v1alpha1.ManifestPath{
						{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "chartsnap-app1",
							JSONPath: []string{
								"/spec/type",
							},
						},
					},
				},
				SnapshotFile:     "__snapshots__/helm_stub_snap_v3.yaml",
				SnapshotVersion:  "", // not specified
				DiffContextLineN: 3,
			}
			res, err := ss.Snap(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Match).To(BeTrueBecause("diff: %s", res.FailureMessage))
		})
	})
})

func TestSnapshotFileName(t *testing.T) {
	type args struct {
		valuesFile string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basename of values file",
			args: args{
				valuesFile: "values.yaml",
			},
			want: "values",
		},
		{
			name: "basename of values file with path",
			args: args{
				valuesFile: "chart/test/values.yaml",
			},
			want: "values",
		},
		{
			name: "default if empty",
			args: args{
				valuesFile: "",
			},
			want: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnapshotFileName(tt.args.valuesFile); got != tt.want {
				t.Errorf("SnapshotFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSnapshotFilePath(t *testing.T) {
	type args struct {
		chartPath  string
		valuesFile string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "same directory with values file",
			args: args{
				chartPath:  "/tmp",
				valuesFile: "chart/test/values.yaml",
			},
			want: "chart/test/__snapshots__/values.snap",
		},
		{
			name: "chart directory with no values file and chart is local",
			args: args{
				chartPath:  "../../example/app1",
				valuesFile: "",
			},
			want: "../../example/app1/__snapshots__/default.snap",
		},
		{
			name: "chart directory with no values file and chart is remote",
			args: args{
				chartPath:  "ingress-nginx/ingress-nginx",
				valuesFile: "",
			},
			want: "__snapshots__/ingress-nginx/__snapshots__/default.snap",
		},
		{
			name: "chart directory with no values file and chart is in OCI registry",
			args: args{
				chartPath:  "oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric",
				valuesFile: "",
			},
			want: "__snapshots__/nginx-gateway-fabric/__snapshots__/default.snap",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultSnapshotFilePath(tt.args.chartPath, tt.args.valuesFile); got != tt.want {
				t.Errorf("DefaultSnapshotFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshotFilePath(t *testing.T) {
	type args struct {
		dir        string
		valuesFile string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ok",
			args: args{
				dir:        "/tmp",
				valuesFile: "values.yaml",
			},
			want: "/tmp/__snapshots__/values.snap",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnapshotFilePath(tt.args.dir, tt.args.valuesFile); got != tt.want {
				t.Errorf("SnapshotFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
