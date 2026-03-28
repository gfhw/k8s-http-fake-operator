{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-http-fake-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
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
Get API Group name
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
Get default values for required parameters
*/}}
{{- define "k8s-http-fake-operator.defaultImageRepository" -}}
{{- default "k8s-http-fake-operator" .Values.image.repository }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultHTTPPort" -}}
{{- default 8080 .Values.service.httpPort }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultHTTPSPort" -}}
{{- default 8443 .Values.service.httpsPort }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultHealthPort" -}}
{{- default 8081 .Values.service.healthPort }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultOperatorHTTPPort" -}}
{{- default 8080 .Values.operator.server.httpPort }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultOperatorHTTPSPort" -}}
{{- default 8443 .Values.operator.server.httpsPort }}
{{- end }}

{{- define "k8s-http-fake-operator.defaultTLSCertSecretName" -}}
{{- default "k8s-http-fake-operator-tls" .Values.tls.certSecretName }}
{{- end }}

{{/*
Validate cluster IP format if provided
*/}}
{{- define "k8s-http-fake-operator.validateClusterIP" -}}
{{- if .Values.service.clusterIP }}
  {{- $clusterIP := .Values.service.clusterIP }}
  {{- $isIPv4 := regexMatch "^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$" $clusterIP }}
  {{- $isIPv6 := regexMatch "^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$" $clusterIP }}
  {{- $isIPv6Compressed := regexMatch "^(([0-9a-fA-F]{1,4}:){0,6}([0-9a-fA-F]{1,4}))?::(([0-9a-fA-F]{1,4}:){0,6}([0-9a-fA-F]{1,4}))?$" $clusterIP }}
  {{- if not (or $isIPv4 $isIPv6 $isIPv6Compressed) }}
  {{- fail "ERROR: service.clusterIP must be a valid IPv4 or IPv6 address. Please set a valid cluster IP in values.yaml." }}
  {{- end }}
{{- end }}
{{- end }}
