package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	alertmanager "github.com/prometheus/alertmanager/template"
)

var logger = log.New(os.Stderr, "", 0)

type Configuration struct {
	PhoneNumber     string
	GroupId         string
	MessageTemplate string
	DataDir         string
	ListenAddress   string
	Verbose         bool
}

type Message struct {
	Status      string
	AlertName   string
	Instance    string
	Severity    string
	Summary     string
	Description string
	Labels      map[string]string
	Annotations map[string]string
}

func NewConfigurationFromEnv() (*Configuration, error) {
	config := Configuration{
		PhoneNumber:     os.Getenv("SIGNAL_RECEIVER_PHONE_NUMBER"),
		GroupId:         os.Getenv("SIGNAL_RECEIVER_GROUP_ID"),
		MessageTemplate: os.Getenv("SIGNAL_RECEIVER_MESSAGE_TEMPLATE"),
		DataDir:         os.Getenv("SIGNAL_RECEIVER_DATA_DIR"),
		ListenAddress:   os.Getenv("SIGNAL_RECEIVER_LISTEN_ADDRESS"),
		Verbose:         os.Getenv("SIGNAL_RECEIVER_VERBOSE") != "",
	}
	if config.PhoneNumber == "" {
		return nil, fmt.Errorf("environment variable SIGNAL_RECEIVER_PHONE_NUMBER empty or undefined")
	}
	if config.GroupId == "" {
		return nil, fmt.Errorf("environment variable SIGNAL_RECEIVER_GROUP_ID empty or undefined")
	}
	if config.MessageTemplate == "" {
		// pls tell me how to do this properly
		config.MessageTemplate = `{{ .Status | ToUpper }}
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
	}
	if config.DataDir == "" {
		config.DataDir = "./data"
	}
	if config.ListenAddress == "" {
		config.ListenAddress = ":9709"
	}
	return &config, nil
}

func (config *Configuration) HandleHealth(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "OK\n")
}

func (config *Configuration) HandleAlert(res http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.WriteError(res, fmt.Errorf("failed to read request body: %w", err), http.StatusBadRequest)
		return
	}
	payload := &alertmanager.Data{}
	err = json.Unmarshal(data, payload)
	if err != nil {
		config.WriteError(res, fmt.Errorf("failed to unmarshal payload: %w", err), http.StatusBadRequest)
		return
	}
	if config.Verbose {
		logger.Printf("webhook triggered with payload: %#v", payload)
	}
	for _, alert := range payload.Alerts {
		msg := NewMessageFromAlert(alert)
		err = msg.Send(config)
		if err != nil {
			config.WriteError(res, fmt.Errorf("failed to send message: %w", err), http.StatusInternalServerError)
			return
		}
	}
}

func (config *Configuration) WriteError(res http.ResponseWriter, err error, code int) {
	logger.Printf("error while handling request: %w", err)
	http.Error(res, err.Error(), code)
}

func NewMessageFromAlert(alert alertmanager.Alert) *Message {
	msg := Message{
		Status:      alert.Status,
		AlertName:   alert.Labels["alertname"],
		Instance:    alert.Labels["instance"],
		Severity:    alert.Labels["severity"],
		Summary:     alert.Annotations["summary"],
		Description: alert.Annotations["description"],
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	for key, value := range alert.Labels {
		if key != "alertname" && key != "instance" && key != "severity" {
			msg.Labels[key] = value
		}
	}
	for key, value := range alert.Annotations {
		if key != "summary" && key != "description" {
			msg.Annotations[key] = value
		}
	}
	return &msg
}

func (msg *Message) Render(config *Configuration) string {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}
	tmpl, err := template.New("message").Funcs(funcMap).Parse(config.MessageTemplate)
	if err != nil {
		logger.Fatal("could not parse message template: ", err)
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, msg)
	if err != nil {
		logger.Fatal("could not instantiate message template: ", err)
	}
	return buffer.String()
}

func (msg *Message) Send(config *Configuration) error {
	text := msg.Render(config)
	if text == "" {
		return fmt.Errorf("refusing to send message: text was empty")
	}
	if config.Verbose {
		logger.Printf("sending message: %s", text)
	}
	var buffer bytes.Buffer
	cmd := exec.Command("signal-cli", "--config", config.DataDir, "--username", config.PhoneNumber, "send", "--group", config.GroupId)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("signal-cli execution error: %s: %w", buffer.String(), err)
	}
	return nil
}

func main() {
	config, err := NewConfigurationFromEnv()
	if err != nil {
		logger.Fatal("could not load configuration: ", err)
	}
	if config.Verbose {
		logger.Printf("configuration: %#v", config)
	}
	http.HandleFunc("/healthz", config.HandleHealth)
	http.HandleFunc("/alert", config.HandleAlert)
	logger.Printf("listening on %s", config.ListenAddress)
	logger.Fatal(http.ListenAndServe(config.ListenAddress, nil))
}
