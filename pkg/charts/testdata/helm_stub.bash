#!/bin/bash



cat <<EOF
this is warning message of helm
---
# Source: app1/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: chartsnap-app1
automountServiceAccountToken: true
---
# Source: app1/templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: app1-cert
  namespace: default
data:
  ca.crt:  $(date +%s | base64)
  tls.crt: $(date +%s | base64)
  tls.key: $(date +%s | base64)
type: kubernetes.io/tls
---
# Source: app1/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: chartsnap-app1
spec:
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: http
  selector:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/name: app1
  type: ClusterIP
---
# Source: app1/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: chartsnap-app1
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: chartsnap
      app.kubernetes.io/name: app1
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: chartsnap
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: app1
        app.kubernetes.io/version: 1.16.0
        helm.sh/chart: app1-0.1.0
    spec:
      containers:
        - image: nginx:1.16.0
          imagePullPolicy: IfNotPresent
          livenessProbe:
            httpGet:
              path: /
              port: http
          name: app1
          ports:
            - containerPort: 80
              name: http
              protocol: TCP
          readinessProbe:
            httpGet:
              path: /
              port: http
          resources: {}
          securityContext: {}
      securityContext: {}
      serviceAccountName: chartsnap-app1
---
# Source: app1/templates/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: chartsnap-app1
spec:
  maxReplicas: 10
  metrics:
    - resource:
        name: cpu
        target:
          averageUtilization: 65
          type: Utilization
      type: Resource
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: chartsnap-app1
---
# Source: app1/templates/tests/test-connection.yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    helm.sh/hook: test
  labels:
    app.kubernetes.io/instance: chartsnap
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: app1
    app.kubernetes.io/version: 1.16.0
    helm.sh/chart: app1-0.1.0
  name: chartsnap-app1-test-connection
spec:
  containers:
    - args:
        - chartsnap-app1:80
      command:
        - wget
      image: busybox
      name: wget
  restartPolicy: Never
EOF