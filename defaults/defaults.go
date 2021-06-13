package defaults

const ListenAddress = ":9709"
const DataDir = "./data"
const MessageTemplate = `{{ .Status | ToUpper }}
{{ .AlertName }}
{{ if .Instance -}}
instance: {{ .Instance }}{{ "\n" }}
{{- end -}}
{{- if .Severity -}}
severity: {{ .Severity }}{{ "\n" }}
{{- end -}}
{{- range $key, $value := .Labels -}}
{{ $key | ToLower }}: {{ $value }}{{ "\n" }}
{{- end -}}
{{- range $key, $value := .Annotations -}}
{{ $key | ToLower }}: {{ $value }}{{ "\n" }}
{{- end -}}
{{- if .Summary -}}
{{ "\n" }}{{ .Summary }}
{{- end -}}
{{- if .Description -}}
{{ "\n" }}{{ .Description }}
{{- end }}`
