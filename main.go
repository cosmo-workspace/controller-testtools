package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/jlandowner/helm-chartsnap/pkg/api/v1alpha1"
	"github.com/jlandowner/helm-chartsnap/pkg/charts"
)

var (
	// goreleaser default https://goreleaser.com/customization/builds/
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
	o       = &option{}
	log     *slog.Logger
	values  []string
	mutex   = &sync.Mutex{}
)

type option struct {
	ReleaseName      string
	Chart            string
	ValuesFile       string
	UpdateSnapshot   bool
	OutputDir        string
	DiffContextLineN int
	FailFast         bool
	Parallelism      int
	ConfigFile       string
	LegacySnapshot   bool
	SnapshotVersion  string

	// Below properties are the same as helm global options
	// They are passed to the plugin as environment variables
	NamespaceFlag string
	DebugFlag     bool
}

func (o *option) Debug() bool {
	helmDebug, err := strconv.ParseBool(os.Getenv("HELM_DEBUG"))
	if err == nil {
		return helmDebug
	}
	return o.DebugFlag
}

func (o *option) Namespace() string {
	helmNamespace := os.Getenv("HELM_NAMESPACE")
	if helmNamespace != "" {
		return helmNamespace
	}
	return o.NamespaceFlag
}

func (o *option) HelmBin() string {
	helmBin := os.Getenv("HELM_BIN")
	if helmBin != "" {
		return helmBin
	}
	return "helm"
}

func (o *option) OK() string {
	if o.UpdateSnapshot {
		return "updated"
	}
	return "matched"
}

// compatibility for --legacy-snapshot flag
func (o *option) snapshotVersion() string {
	// use v1 snapshot format if legacy snapshot format is enabled
	if o.LegacySnapshot {
		return charts.SnapshotVersionV1
	} else {
		return o.SnapshotVersion
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "chartsnap -c CHART",
		Short: "Snapshot testing tool for Helm charts",
		Long: `
Snapshot testing tool like Jest for Helm charts.

You can create test cases as a variation of Values files of your chart.
` + "`" + `__snapshot__` + "`" + ` directory is created in the same directory as test Values files.

In addition, chartsnap support preventing mismatched snapshots by Helm functions.
You can specify the paths of dynamic values in the generated YAML using [JSONPath](https://datatracker.ietf.org/doc/html/rfc6901).

` + "```" + `yaml
# dynamicFields defines values that are dynamically generated by Helm function like 'randAlphaNum'
# https://helm.sh/docs/chart_template_guide/function_list/#randalphanum-randalpha-randnumeric-and-randascii
# Replace outputs with fixed values to prevent unmatched outputs at each snapshot.
dynamicFields:
  - apiVersion: v1
    kind: Secret
    name: cosmo-auth-env
    jsonPath:
      - /data/COOKIE_HASHKEY
      - /data/COOKIE_BLOCKKEY
      - /data/COOKIE_HASHKEY
      - /data/COOKIE_SESSION_NAME
` + "```" + `

Place it as a '.chartsnap.yaml' file within your test values file directory.

See the repository for the full documentation.
https://github.com/jlandowner/helm-chartsnap.git

MIT 2023 jlandowner/helm-chartsnap
`,
		Example: `
  # Snapshot with default values:
  chartsnap -c YOUR_CHART
  
  # Update snapshot files:
  chartsnap -c YOUR_CHART -u

  # Snapshot with test case values:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILE
  
  # Snapshot all test cases:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILES_DIRECTOY
  
  # Set additional args or flags for the 'helm template' command:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILE -- --skip-tests

  # Snapshot remote chart in Helm repository:
  chartsnap -c CHART_NAME -f YOUR_VALUES_FILE -- --repo HELM_REPO_URL

  # Snapshot ingress-nginx (https://kubernetes.github.io/ingress-nginx/) helm chart for a specific version with your value file:
  chartsnap -c ingress-nginx -f YOUR_VALUES_FILE -- --repo https://kubernetes.github.io/ingress-nginx --namespace kube-system --version 4.8.3

  # Snapshot cilium (https://cilium.io) helm chart with default value and set flags:
  chartsnap -c cilium -- --repo https://helm.cilium.io --namespace kube-system --set hubble.relay.enabled=true --set hubble.ui.enabled=true

  # Snapshot charts in OCI registry
  chartsnap -c oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -n nginx-gateway

  # Output with no colors:
  NO_COLOR=1 chartsnap -c YOUR_CHART`,
		Version: fmt.Sprintf("version=%s commit=%s date=%s", version, commit, date),
		RunE:    run,
		PreRunE: prerun,
	}
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.PersistentFlags().BoolVar(&o.DebugFlag, "debug", false, "debug mode")
	rootCmd.PersistentFlags().BoolVarP(&o.UpdateSnapshot, "update-snapshot", "u", false, "update snapshot mode")
	rootCmd.PersistentFlags().StringVarP(&o.Chart, "chart", "c", "", "path to the chart directory. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES' as 'CHART'")
	if err := rootCmd.MarkPersistentFlagDirname("chart"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("chart"); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().StringVar(&o.ReleaseName, "release-name", "chartsnap", "release name. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES' as 'RELEASE_NAME'")
	rootCmd.PersistentFlags().StringVarP(&o.NamespaceFlag, "namespace", "n", "default", "namespace. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES --namespace NAMESPACE' as 'NAMESPACE'")
	rootCmd.PersistentFlags().StringVarP(&o.ValuesFile, "values", "f", "", "path to a test values file or directory. if the directory is set, all test files are tested. if empty, default values are used. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES' as 'VALUES'")
	rootCmd.PersistentFlags().StringVarP(&o.OutputDir, "output-dir", "o", "", "directory which is __snapshot__ directory is created. (default: values file directory if --values is set; chart directory if chart is local; else current directory)")
	if err := rootCmd.MarkPersistentFlagDirname("output-dir"); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().IntVarP(&o.DiffContextLineN, "ctx-lines", "N", 3, "number of lines to show in diff output. 0 for full output")
	rootCmd.PersistentFlags().BoolVar(&o.FailFast, "failfast", false, "fail once any test case failed")
	rootCmd.PersistentFlags().IntVar(&o.Parallelism, "parallelism", -1, "test concurrency if taking multiple snapshots for a test value file directory. default is unlimited")
	rootCmd.PersistentFlags().StringVar(&o.ConfigFile, "config-file", ".chartsnap.yaml", "config file name or path, which defines snapshot behavior e.g. dynamic fields")
	if err := rootCmd.MarkPersistentFlagFilename("config-file"); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().BoolVar(&o.LegacySnapshot, "legacy-snapshot", false, "use toml-based legacy snapshot format")
	rootCmd.PersistentFlags().StringVar(&o.SnapshotVersion, "snapshot-version", "", "use a specific snapshot format version. v1, v2, v3 are supported. (default: latest)")

	if err := rootCmd.Execute(); err != nil {
		slog.New(slogHandler()).Error(err.Error())
		os.Exit(1)
	}
}

func slogHandler() slog.Handler {
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: func() slog.Leveler {
			if o.Debug() {
				return slog.LevelDebug
			}
			return slog.LevelInfo
		}(),
	})
}

func prerun(cmd *cobra.Command, args []string) error {
	if o.Chart == "" {
		// show help message when executed without any args (meaning required --chart flag is not set)
		return cmd.Help()
	}
	return nil
}

func loadSnapshotConfig(file string, cfg *v1alpha1.SnapshotConfig) error {
	err := v1alpha1.FromFile(file, cfg)
	if err != nil && !os.IsNotExist(err) {
		if o.FailFast {
			return fmt.Errorf("failed to load snapshot config: %w", err)
		} else {
			log.Error("WARNING: failed to load snapshot config", "path", file, "err", err)
		}
	}
	log.Debug("snapshot config", "cfg", cfg)
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	log = slog.New(slogHandler())
	log.Debug("options", printOptions(*o)...)
	log.Debug("args", "args", args)
	charts.SetLogger(log)

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "HELM_") {
			e := strings.Split(v, "=")
			log.Debug("helm env", "key", e[0], "value", e[1])
		}
	}

	var cfg v1alpha1.SnapshotConfig
	if _, err := os.Stat(o.ConfigFile); err == nil {
		if err := loadSnapshotConfig(o.ConfigFile, &cfg); err != nil {
			return err
		}
	}

	if o.ValuesFile == "" {
		values = []string{""}
	} else {
		stat, err := os.Stat(o.ValuesFile)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("values file '%s' not found", o.ValuesFile)
			}
			return fmt.Errorf("failed to stat values file %s: %w", o.ValuesFile, err)
		}

		if stat.IsDir() {
			// get all values files in the directory
			files, err := os.ReadDir(o.ValuesFile)
			if err != nil {
				return fmt.Errorf("failed to read values file directory: %w", err)
			}
			values = make([]string, 0)
			for _, f := range files {
				// pick config file in a test values directory
				if f.Name() == o.ConfigFile {
					if err = loadSnapshotConfig(path.Join(o.ValuesFile, f.Name()), &cfg); err != nil {
						return err
					}
					continue
				}

				// read test values files (only *.yaml)
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
					values = append(values, path.Join(o.ValuesFile, f.Name()))
				}
			}
		} else {
			values = []string{o.ValuesFile}

			// load .chartsnap config in the base directory if exist
			dirCfg := path.Join(path.Dir(o.ValuesFile), o.ConfigFile)
			if _, err := os.Stat(dirCfg); err == nil {
				if err := loadSnapshotConfig(dirCfg, &cfg); err != nil {
					return err
				}
			}
		}
	}

	eg, ctx := errgroup.WithContext(cmd.Context())
	if !o.FailFast {
		// not cancel ctx even if some case failed
		ctx = cmd.Context()
	}
	// limit concurrency to o.Parallelism
	eg.SetLimit(o.Parallelism)
	if o.Debug() {
		// limit concurrency to 1 for debugging.
		eg.SetLimit(1)
	}
	for _, v := range values {
		ht := charts.HelmTemplateCmdOptions{
			HelmPath:       o.HelmBin(),
			ReleaseName:    o.ReleaseName,
			Namespace:      o.Namespace(),
			Chart:          o.Chart,
			ValuesFile:     v,
			AdditionalArgs: args,
		}
		bannerPrintln("RUNS",
			fmt.Sprintf("Snapshot testing chart=%s values=%s", ht.Chart, ht.ValuesFile), 0, color.BgBlue)
		eg.Go(func() error {
			var snapshotFilePath string
			if o.OutputDir != "" {
				snapshotFilePath = charts.SnapshotFilePath(o.OutputDir, ht.ValuesFile)
			} else {
				snapshotFilePath = charts.DefaultSnapshotFilePath(ht.Chart, ht.ValuesFile)
			}

			snapshotter := charts.ChartSnapshotter{
				HelmTemplateCmdOptions: ht,
				SnapshotConfig:         cfg,
				SnapshotFile:           snapshotFilePath,
				SnapshotVersion:        o.snapshotVersion(),
				DiffContextLineN:       o.DiffContextLineN,
				UpdateSnapshot:         o.UpdateSnapshot,
				HeaderVersion:          version,
			}
			result, err := snapshotter.Snap(ctx)
			if err != nil {
				bannerPrintln("FAIL", fmt.Sprintf("chart=%s values=%s err=%v snapshot_version=%s", ht.Chart, ht.ValuesFile, snapshotter.SnapshotVersion, err), color.FgRed, color.BgRed)
				return fmt.Errorf("failed to get snapshot chart=%s values=%s: %w", ht.Chart, ht.ValuesFile, err)
			}
			if !result.Match {
				bannerPrintln("FAIL", fmt.Sprintf("Snapshot does not match chart=%s values=%s snapshot_version=%s", ht.Chart, ht.ValuesFile, snapshotter.SnapshotVersion), color.FgRed, color.BgRed)
				fmt.Println(result.FailureMessage)
				return fmt.Errorf("snapshot does not match chart=%s values=%s", ht.Chart, ht.ValuesFile)
			}
			bannerPrintln("PASS", fmt.Sprintf("Snapshot %s chart=%s values=%s snapshot_version=%s", o.OK(), ht.Chart, ht.ValuesFile, snapshotter.SnapshotVersion), color.FgGreen, color.BgGreen)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	bannerPrintln("PASS", fmt.Sprintf("All snapshots %s", o.OK()), color.FgGreen, color.BgGreen)

	return nil
}

func bannerPrintln(banner string, message string, fgColor color.Attribute, bgColor color.Attribute) {
	mutex.Lock()
	defer mutex.Unlock()
	color.New(color.FgWhite, bgColor).Printf(" %s ", banner)
	color.New(fgColor).Printf(" %s\n", message)
}

func printOptions(o option) []any {
	rv := reflect.ValueOf(o)
	rt := rv.Type()
	options := make([]any, rt.NumField()*2)

	for i := 0; i < rt.NumField(); i++ {
		options[i*2] = rt.Field(i).Name
		options[i*2+1] = rv.Field(i).Interface()
	}

	return options
}
