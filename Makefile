all: build

GO ?= go1.22.2

go:
	-go install golang.org/dl/$(GO)@latest
	$(GO) download
	rm -f $$(dirname $$(which $(GO)))/go
	ln -s $$(which $(GO)) $$(dirname $$(which $(GO)))/go
	go version

helm:
	curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

goreleaser:
	$(GO) install github.com/goreleaser/goreleaser@latest

.PHONY: build
build: goreleaser
	goreleaser build --single-target --snapshot --clean --skip=before

.PHONY: test
test:
	$(GO) test ./... -cover cover.out

HELM_PLUGIN_PATH := $(shell helm env | grep HELM_PLUGINS | cut -d= -f2)

.PHONY: integ-test
integ-test: install-dev-bin
	helm chartsnap --chart example/app1 -f example/app1/test/test_ingress_enabled.yaml --namespace default $(ARGS)
	helm chartsnap --chart example/app1 -f example/app1/test/ --namespace default $(ARGS)
	helm chartsnap --chart oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -f example/remote/nginx-gateway-fabric.values.yaml $(ARGS) -- --namespace nginx-gateway $(EXTRA_ARGS)
	helm chartsnap --chart cilium -f example/remote/cilium.values.yaml $(ARGS) -- --namespace kube-system --repo https://helm.cilium.io $(EXTRA_ARGS)
	helm chartsnap --chart ingress-nginx -f example/remote/ingress-nginx.values.yaml $(ARGS) -- --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --skip-tests $(EXTRA_ARGS)

.PHONY: integ-test-fail
integ-test-fail: install-dev-bin
	-helm chartsnap --chart example/app1 --namespace default $(ARGS)
	-helm chartsnap --chart example/app1 --namespace default -f example/app1/testfail/test_ingress_enabled.yaml $(ARGS)
	-helm chartsnap --chart example/app1 --namespace default -f example/app1/testfail/ $(ARGS)

.PHONY: update-versions
update-versions:
	sed -i.bk 's/version: .*/version: $(VERSION)/' plugin.yaml

.PHONY: install-dev-bin
install-dev-bin: build
	-helm plugin install https://github.com/jlandowner/helm-chartsnap
	cp ./dist/chartsnap_*/chartsnap $(HELM_PLUGIN_PATH)/helm-chartsnap/bin/
	helm chartsnap --version

.PHONY: helm-template-help-snapshot
helm-template-help-snapshot:
	cd hack/helm-template-help-snapshot; $(GO) run main.go

.PHONY: helm-template-diff
helm-template-diff:
	cd hack/helm-template-diff; $(GO) run main.go
