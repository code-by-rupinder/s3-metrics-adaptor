{{/*
Expand the name of the chart.
*/}}
{{- define "s3-metrics-adapter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "s3-metrics-adapter.fullname" -}}
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
{{- define "s3-metrics-adapter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "s3-metrics-adapter.labels" -}}
helm.sh/chart: {{ include "s3-metrics-adapter.chart" . }}
{{ include "s3-metrics-adapter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "s3-metrics-adapter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "s3-metrics-adapter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "s3-metrics-adapter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "s3-metrics-adapter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the AWS credentials secret
*/}}
{{- define "s3-metrics-adapter.awsSecretName" -}}
{{- if .Values.awsCredentials.secretName }}
{{- .Values.awsCredentials.secretName }}
{{- else }}
{{- printf "%s-aws-credentials" (include "s3-metrics-adapter.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Determine which AWS credentials secret to use
*/}}
{{- define "s3-metrics-adapter.awsCredentialsSecretName" -}}
{{- if .Values.awsCredentials.existingSecret.name }}
{{- .Values.awsCredentials.existingSecret.name }}
{{- else if .Values.awsCredentials.create }}
{{- include "s3-metrics-adapter.awsSecretName" . }}
{{- end }}
{{- end }}

{{/*
Generate AWS environment variables from credentials
*/}}
{{- define "s3-metrics-adapter.awsEnvVars" -}}
{{- $secretName := include "s3-metrics-adapter.awsCredentialsSecretName" . }}
{{- if $secretName }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ .Values.awsCredentials.existingSecret.accessKeyIdKey | default "AWS_ACCESS_KEY_ID" }}
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ .Values.awsCredentials.existingSecret.secretAccessKeyKey | default "AWS_SECRET_ACCESS_KEY" }}
{{- if or .Values.awsCredentials.sessionToken .Values.awsCredentials.existingSecret.sessionTokenKey }}
- name: AWS_SESSION_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ .Values.awsCredentials.existingSecret.sessionTokenKey | default "AWS_SESSION_TOKEN" }}
      optional: true
{{- end }}
- name: AWS_REGION
  valueFrom:
    secretKeyRef:
      name: {{ $secretName }}
      key: {{ .Values.awsCredentials.existingSecret.regionKey | default "AWS_REGION" }}
{{- end }}
{{- end }}
