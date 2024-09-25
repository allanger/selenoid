{{/*
Expand the name of the chart.
*/}}
{{- define "selenoid-kube.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "selenoid-kube.fullname" -}}
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
{{- define "selenoid-kube.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "selenoid-kube.labels" -}}
helm.sh/chart: {{ include "selenoid-kube.chart" . }}
{{ include "selenoid-kube.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "selenoid-kube.selectorLabels" -}}
app.kubernetes.io/name: {{ include "selenoid-kube.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "selenoid-kube.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "selenoid-kube.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Common labels for UI
*/}}
{{- define "selenoidUI.labels" -}}
helm.sh/chart: {{ include "selenoid-kube.chart" . }}
{{ include "selenoidUI.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}



{{/*
Selector labels for the UI
*/}}
{{- define "selenoidUI.selectorLabels" -}}
app.kubernetes.io/name: {{ include "selenoid-kube.name" . }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
