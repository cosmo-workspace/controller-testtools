package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/jlandowner/helm-chartsnap/pkg/charts"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	// goreleaser default https://goreleaser.com/customization/builds/
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
	o       = &option{}
	log     *slog.Logger
	values  []string
)

type option struct {
	ReleaseName      string
	Chart            string
	ValuesFile       string
	UpdateSnapshot   bool
	OutputDir        string
	DiffContextLineN int
	FailOnce         bool

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

func main() {
	rootCmd := &cobra.Command{
		Use:   "chartsnap -c CHART",
		Short: "Snapshot testing tool for Helm charts",
		Long: `
Snapshot testing tool like Jest for Helm charts.

You can create test cases as a variation of Values files of your chart.
` + "`" + `__snapshot__` + "`" + ` directory is created in the same directory as test Values files.
In addition, Values files can have a ` + "`" + `testSpec` + "`" + ` property that can detail or control the test case.

` + "```" + `yaml
testSpec:
  # desc is a description for the set of values
  desc: only required values and the rest is default
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

# Others can be any your chart value.
# ...
` + "```" + `

See the repository for full documentation.
https://github.com/jlandowner/helm-chartsnap.git

MIT 2023 jlandowner/helm-chartsnap
`,
		Example: `
  # Snapshot with defualt values:
  chartsnap -c YOUR_CHART
  
  # Update snapshot files:
  chartsnap -c YOUR_CHART -u

  # Snapshot with test case values:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILE
  
  # Snapshot all test cases:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILES_DIRECTOY
  
  # Set addtional args or flags for 'helm template' command:
  chartsnap -c YOUR_CHART -f YOUR_TEST_VALUES_FILE -- --skip-tests

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
	rootCmd.PersistentFlags().StringVar(&o.NamespaceFlag, "namespace", "default", "namespace. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES --namespace NAMESPACE' as 'NAMESPACE'")
	rootCmd.PersistentFlags().StringVarP(&o.ValuesFile, "values", "f", "", "path to a test values file or directory. if directroy is set, all test files are tested. if empty, default values are used. this flag is passed to 'helm template RELEASE_NAME CHART --values VALUES' as 'VALUES'")
	rootCmd.PersistentFlags().StringVarP(&o.OutputDir, "output-dir", "o", "", "directory which is __snapshot__ directory is created. (default: values file directory if --values is set; chart directory if chart is local; else current directory)")
	if err := rootCmd.MarkPersistentFlagDirname("output-dir"); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().IntVarP(&o.DiffContextLineN, "ctx-lines", "N", 3, "number of lines to show in diff output. 0 for full output")
	rootCmd.PersistentFlags().BoolVar(&o.FailOnce, "fail-once", false, "fail once any test case failed")

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

func run(cmd *cobra.Command, args []string) error {
	log = slog.New(slogHandler())
	log.Debug("options", printOptions(*o)...)
	log.Debug("args", "args", args)

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "HELM_") {
			e := strings.Split(v, "=")
			log.Debug("helm env", "key", e[0], "value", e[1])
		}
	}

	if o.ValuesFile == "" {
		values = []string{""}
	} else {
		if s, err := os.Stat(o.ValuesFile); os.IsNotExist(err) {
			return fmt.Errorf("values file '%s' not found", o.ValuesFile)
		} else if s.IsDir() {
			// get all values files in the directory
			files, err := os.ReadDir(o.ValuesFile)
			if err != nil {
				return fmt.Errorf("failed to read values file directory: %w", err)
			}
			values = make([]string, 0)
			for _, f := range files {
				// read only *.yaml
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".yaml") {
					values = append(values, path.Join(o.ValuesFile, f.Name()))
				}
			}
		} else {
			values = []string{o.ValuesFile}
		}
	}

	eg, ctx := errgroup.WithContext(cmd.Context())
	if !o.FailOnce {
		// not cancel ctx even if some case failed
		ctx = cmd.Context()
	}
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
		charts.SetLogger(log)
		bannerPrintln("RUNS",
			fmt.Sprintf("Snapshot testing chart=%s values=%s", ht.Chart, ht.ValuesFile), 0, color.BgBlue)
		eg.Go(func() error {
			var snapshotFilePath string
			if o.OutputDir != "" {
				snapshotFilePath = charts.SnapshotFilePath(o.OutputDir, ht.ValuesFile)
			} else {
				snapshotFilePath = charts.DefaultSnapshotFilePath(ht.Chart, ht.ValuesFile)
			}

			_, err := os.Stat(snapshotFilePath)
			if err == nil {
				log.Debug("snapshot file already exists", "path", snapshotFilePath)
			} else if os.IsNotExist(err) {
				log.Debug("snapshot file does not exist", "path", snapshotFilePath)
			}

			if o.UpdateSnapshot {
				err := os.Remove(snapshotFilePath)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to replace snapshot file: %w", err)
				}
			}

			opts := charts.ChartSnapOptions{
				HelmTemplateCmdOptions: ht,
				SnapshotFile:           snapshotFilePath,
				DiffContextLineN:       o.DiffContextLineN,
			}
			matched, failureMessage, err := charts.Snap(ctx, opts)
			if err != nil {
				bannerPrintln("FAIL", fmt.Sprintf("chart=%s values=%s err=%v", ht.Chart, ht.ValuesFile, err), color.FgRed, color.BgRed)
				return fmt.Errorf("failed to get snapshot chart=%s values=%s: %w", ht.Chart, ht.ValuesFile, err)
			}
			if !matched {
				bannerPrintln("FAIL", fmt.Sprintf("Snapshot does not match chart=%s values=%s", ht.Chart, ht.ValuesFile), color.FgRed, color.BgRed)
				fmt.Println(failureMessage)
				return fmt.Errorf("snapshot does not match chart=%s values=%s", ht.Chart, ht.ValuesFile)
			}
			bannerPrintln("PASS", fmt.Sprintf("Snapshot matched chart=%s values=%s", ht.Chart, ht.ValuesFile), color.FgGreen, color.BgGreen)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}
	bannerPrintln("PASS", "All snapshot matched", color.FgGreen, color.BgGreen)

	return nil
}

func bannerPrintln(banner string, message string, fgColor color.Attribute, bgColor color.Attribute) {
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
