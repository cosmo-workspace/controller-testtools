package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"

	"github.com/jlandowner/helm-chartsnap/pkg/snap"
)

func main() {
	cases := []struct {
		snap string
		helm string
	}{
		{
			snap: "../../example/app1/__snapshots__/default.snap",
			helm: "helm template chartsnap ../../example/app1 -n default",
		},
		{
			snap: "../../example/app1/test/__snapshots__/test_ingress_enabled.snap",
			helm: "helm template chartsnap ../../example/app1 -f ../../example/app1/test/test_ingress_enabled.yaml -n default",
		},
		{
			snap: "../../example/app1/test/__snapshots__/test_hpa_enabled.snap",
			helm: "helm template chartsnap ../../example/app1 -f ../../example/app1/test/test_hpa_enabled.yaml -n default",
		},
		{
			snap: "../../example/app1/test/__snapshots__/test_certmanager_enabled.snap",
			helm: "helm template chartsnap ../../example/app1 -f ../../example/app1/test/test_certmanager_enabled.yaml -n default",
		},
		{
			snap: "../../example/remote/__snapshots__/nginx-gateway-fabric.values.snap",
			helm: "helm template chartsnap oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -f ../../example/remote/nginx-gateway-fabric.values.yaml -n nginx-gateway",
		},
		{
			snap: "../../example/remote/__snapshots__/cilium.values.snap",
			helm: "helm template chartsnap cilium -f ../../example/remote/cilium.values.yaml -n kube-system --repo https://helm.cilium.io",
		},
		{
			snap: "../../example/remote/__snapshots__/ingress-nginx.values.snap",
			helm: "helm template chartsnap ingress-nginx -f ../../example/remote/ingress-nginx.values.yaml --repo https://kubernetes.github.io/ingress-nginx -n ingress-nginx --skip-tests",
		},
	}

	for _, c := range cases {
		out := execute("sh", "-c", c.helm)

		re := regexp.MustCompile(`LS0tLS1CRUdJTiB[^ ]*`)
		replaced := re.ReplaceAll(out, []byte("REPLACEMENT"))

		f, err := os.CreateTemp("", "helm-template-diff")
		if err != nil {
			slog.Error("create temp file error", "err", err)
			os.Exit(9)
		}
		defer os.Remove(f.Name())

		_, err = f.Write(replaced)
		if err != nil {
			slog.Error("write temp file error", "err", err)
			os.Exit(9)
		}

		snapshot(fmt.Sprintf("sdiff %s <(%s)", c.snap, c.helm), string(execute("sdiff", c.snap, f.Name())))
	}
}

func execute(cmd ...string) []byte {
	out, _ := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	return out
}

func snapshot(id, data string) {
	s := snap.SnapshotMatcher("helm-template.snap", snap.WithSnapshotID(id))
	match, err := s.Match(data)

	if err != nil {
		slog.Error("snapshot error", "err", err)
		os.Exit(9)
	}
	if !match {
		slog.Error(s.FailureMessage(nil))
		os.Exit(1)
	}
}
