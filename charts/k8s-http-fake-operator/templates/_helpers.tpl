{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-http-fake-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8s-http-fake-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "k8s-http-fake-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-http-fake-operator.labels" -}}
helm.sh/chart: {{ include "k8s-http-fake-operator.chart" . }}
{{ include "k8s-http-fake-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-http-fake-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-http-fake-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8s-http-fake-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "k8s-http-fake-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Default HTTP port
*/}}
{{- define "k8s-http-fake-operator.defaultHTTPPort" -}}
{{- default 8080 .Values.operator.server.httpPort }}
{{- end }}

{{/*
Default HTTPS port
*/}}
{{- define "k8s-http-fake-operator.defaultHTTPSPort" -}}
{{- default 8443 .Values.operator.server.httpsPort }}
{{- end }}

{{/*
Default health port
*/}}
{{- define "k8s-http-fake-operator.defaultHealthPort" -}}
{{- default 8081 .Values.operator.server.healthPort }}
{{- end }}

{{/*
Default operator HTTP port (alias for configmap compatibility)
*/}}
{{- define "k8s-http-fake-operator.defaultOperatorHTTPPort" -}}
{{- include "k8s-http-fake-operator.defaultHTTPPort" . }}
{{- end }}

{{/*
Default operator HTTPS port (alias for configmap compatibility)
*/}}
{{- define "k8s-http-fake-operator.defaultOperatorHTTPSPort" -}}
{{- include "k8s-http-fake-operator.defaultHTTPSPort" . }}
{{- end }}

{{/*
Default operator health port (alias for configmap compatibility)
*/}}
{{- define "k8s-http-fake-operator.defaultOperatorHealthPort" -}}
{{- include "k8s-http-fake-operator.defaultHealthPort" . }}
{{- end }}

{{/*
Default API group
*/}}
{{- define "k8s-http-fake-operator.apiGroup" -}}
{{- if .Values.apiGroup.fullName }}
{{- .Values.apiGroup.fullName }}
{{- else if .Values.apiGroup.suffix }}
{{- printf "httpteststub.%s.com" .Values.apiGroup.suffix }}
{{- else }}
{{- "httpteststub.example.com" }}
{{- end }}
{{- end }}

{{/*
Default image repository
*/}}
{{- define "k8s-http-fake-operator.defaultImageRepository" -}}
{{- default "k8s-http-fake-operator" .Values.image.repository }}
{{- end }}

{{/*
Default TLS certificate secret name
*/}}
{{- define "k8s-http-fake-operator.defaultTLSCertSecretName" -}}
{{- if .Values.tls.certSecretName }}
{{- .Values.tls.certSecretName }}
{{- else }}
{{- printf "%s-tls" (include "k8s-http-fake-operator.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Validate ClusterIP configuration
*/}}
{{- define "k8s-http-fake-operator.validateClusterIP" -}}
{{- if .Values.service.clusterIP }}
{{- $clusterIP := .Values.service.clusterIP }}
{{- if ne $clusterIP "None" }}
{{- $isIPv4 := regexMatch "^([0-9]{1,3}\\.){3}[0-9]{1,3}$" $clusterIP }}
{{- $isIPv6 := regexMatch "^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$" $clusterIP }}
{{- $isIPv6Compressed := regexMatch "^([0-9a-fA-F]{1,4}:){0,7}:([0-9a-fA-F]{1,4}:){0,7}[0-9a-fA-F]{1,4}$" $clusterIP }}
{{- if not (or $isIPv4 $isIPv6 $isIPv6Compressed) }}
{{- fail "ERROR: service.clusterIP must be a valid IPv4 or IPv6 address. Please set a valid cluster IP in values.yaml." }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Generate TLS certificate using pre-generated cert for compatibility
*/}}
{{- define "k8s-http-fake-operator.genCert" -}}
{{- $cert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMrekNDQWVPZ0F3SUJBZ0lKQUxETlVYTXdDZ1lJS29aSXpqMEVBd0l3RlRFVE1CRUdBMVVFQXd3S2EzVmkKWlhKdVpYUmxjekFlRncweU5EQXhNREV3TURBd01EQmFGdzB5TlRBeE1ERXdNREF3TURCYU1CVXhFekFSQmdOVgpCQU1NQ210MVltVnlibVYwWlhNd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUMKdGVzdC1jZXJ0aWZpY2F0ZQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==" -}}
{{ $cert }}
{{- end }}

{{/*
Generate TLS key using pre-generated key for compatibility
*/}}
{{- define "k8s-http-fake-operator.genKey" -}}
{{- $key := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQ3Rlc3Qta2V5Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K" -}}
{{ $key }}
{{- end }}
